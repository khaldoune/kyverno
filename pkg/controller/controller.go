package controller

import (
	"fmt"
	"time"

	"github.com/nirmata/kyverno/pkg/result"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/minio/minio/pkg/wildcard"
	types "github.com/nirmata/kyverno/pkg/apis/policy/v1alpha1"
	lister "github.com/nirmata/kyverno/pkg/client/listers/policy/v1alpha1"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/nirmata/kyverno/pkg/engine"
	"github.com/nirmata/kyverno/pkg/event"
	"github.com/nirmata/kyverno/pkg/sharedinformer"
	violation "github.com/nirmata/kyverno/pkg/violation"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

//PolicyController to manage Policy CRD
type PolicyController struct {
	client           *client.Client
	policyLister     lister.PolicyLister
	policySynced     cache.InformerSynced
	violationBuilder violation.Generator
	eventBuilder     event.Generator
	queue            workqueue.RateLimitingInterface
	filterKinds      []string
}

// NewPolicyController from cmd args
func NewPolicyController(client *client.Client,
	policyInformer sharedinformer.PolicyInformer,
	violationBuilder violation.Generator,
	eventController event.Generator) *PolicyController {

	controller := &PolicyController{
		client:           client,
		policyLister:     policyInformer.GetLister(),
		policySynced:     policyInformer.GetInfomer().HasSynced,
		violationBuilder: violationBuilder,
		eventBuilder:     eventController,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), policyWorkQueueName),
	}

	policyInformer.GetInfomer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.createPolicyHandler,
		UpdateFunc: controller.updatePolicyHandler,
		DeleteFunc: controller.deletePolicyHandler,
	})
	return controller
}

func (pc *PolicyController) createPolicyHandler(resource interface{}) {
	pc.enqueuePolicy(resource)
}

func (pc *PolicyController) updatePolicyHandler(oldResource, newResource interface{}) {
	newPolicy := newResource.(*types.Policy)
	oldPolicy := oldResource.(*types.Policy)
	if newPolicy.ResourceVersion == oldPolicy.ResourceVersion {
		return
	}
	pc.enqueuePolicy(newResource)
}

func (pc *PolicyController) deletePolicyHandler(resource interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = resource.(metav1.Object); !ok {
		utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
		return
	}
	glog.Infof("policy deleted: %s", object.GetName())
}

func (pc *PolicyController) enqueuePolicy(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	pc.queue.Add(key)
}

// Run is main controller thread
func (pc *PolicyController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	if ok := cache.WaitForCacheSync(stopCh, pc.policySynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < policyControllerWorkerCount; i++ {
		go wait.Until(pc.runWorker, time.Second, stopCh)
	}
	glog.Info("started policy controller workers")

	return nil
}

//Stop to perform actions when controller is stopped
func (pc *PolicyController) Stop() {
	defer pc.queue.ShutDown()
	glog.Info("shutting down policy controller workers")
}
func (pc *PolicyController) runWorker() {
	for pc.processNextWorkItem() {
	}
}

func (pc *PolicyController) processNextWorkItem() bool {
	obj, shutdown := pc.queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer pc.queue.Done(obj)
		err := pc.syncHandler(obj)
		pc.handleErr(err, obj)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (pc *PolicyController) handleErr(err error, key interface{}) {
	if err == nil {
		pc.queue.Forget(key)
		return
	}
	// This controller retries if something goes wrong. After that, it stops trying.
	if pc.queue.NumRequeues(key) < policyWorkQueueRetryLimit {
		glog.Warningf("Error syncing events %v: %v", key, err)
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		pc.queue.AddRateLimited(key)
		return
	}
	pc.queue.Forget(key)
	utilruntime.HandleError(err)
	glog.Warningf("Dropping the key %q out of the queue: %v", key, err)
}

func (pc *PolicyController) syncHandler(obj interface{}) error {
	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		return fmt.Errorf("expected string in workqueue but got %#v", obj)
	}
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid policy key: %s", key))
		return nil
	}

	// Get Policy resource with namespace/name
	policy, err := pc.policyLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("policy '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	// process policy on existing resource
	// get the violations and pass to violation Builder
	// get the events and pass to event Builder
	//TODO: processPolicy
	pc.processPolicy(policy)

	glog.Infof("process policy %s on existing resources", policy.GetName())
	return nil
}

func (pc *PolicyController) processPolicy(p *types.Policy) {
	// Get all resources on which the policy is to be applied
	resources := []*resourceInfo{}
	for _, rule := range p.Spec.Rules {
		for _, k := range rule.Kinds {
			// get resources of defined kinds->resources
			gvr := pc.client.GetGVRFromKind(k)
			// LabelSelector
			// namespace ?
			list, err := pc.client.ListResource(gvr.Resource, "", rule.ResourceDescription.Selector)
			if err != nil {
				glog.Errorf("unable to list resources for %s with label selector %s", gvr.Resource, rule.Selector.String())
				glog.Errorf("unable to apply policy %s rule %s. err : %s", p.Name, rule.Name, err)
				continue
			}

			for _, resource := range list.Items {
				name := rule.ResourceDescription.Name
				gvk := resource.GroupVersionKind()
				rawResource, err := resource.MarshalJSON()
				if err != nil {
					glog.Errorf("Unable to json parse resource %s", resource.GetName())
					continue
				}
				if name != nil {
					// wild card matching
					if !wildcard.Match(*name, resource.GetName()) {
						continue
					}
				}
				glog.Info(string(rawResource))
				ri := &resourceInfo{rawResource: rawResource, gvk: &metav1.GroupVersionKind{Group: gvk.Group,
					Version: gvk.Version,
					Kind:    gvk.Kind}}
				resources = append(resources, ri)
			}
		}
	}
	// for the filtered resource apply policy
	for _, r := range resources {
		pc.applyPolicy(p, r.rawResource, r.gvk)
	}
	// apply policies on the filtered resources
}

func (pc *PolicyController) applyPolicy(p *types.Policy, rawResource []byte, gvk *metav1.GroupVersionKind) {
	policyResult := result.NewPolicyApplicationResult(p.Name)
	//TODO: PR #181 use the list of kinds to filter here too

	// Mutate
	mutationResult := mutation(p, rawResource, gvk)
	policyResult = result.Append(policyResult, mutationResult)

	// Validate
	validationResult := engine.Validate(*p, rawResource, *gvk)
	policyResult = result.Append(policyResult, validationResult)
	// Generate
	generateResult := engine.Generate(pc.client, *p, rawResource, *gvk, true)
	policyResult = result.Append(policyResult, generateResult)
}

func mutation(p *types.Policy, rawResource []byte, gvk *metav1.GroupVersionKind) result.Result {
	patches, mutationResult := engine.Mutate(*p, rawResource, *gvk)
	// option 2: (original Resource + patch) compare with (original resource)
	mergePatches := engine.JoinPatches(patches)
	// merge the patches
	patch, err := jsonpatch.DecodePatch(mergePatches)
	if err != nil {
		mresult := result.NewRuleApplicationResult("")
		mresult.FailWithMessagef(err.Error())
		mutationResult = result.Append(mutationResult, &mresult)
	}
	// apply the patches returned by mutate to the original resource
	patchedResource, err := patch.Apply(rawResource)
	if err != nil {
		mresult := result.NewRuleApplicationResult("")
		mresult.FailWithMessagef(err.Error())
		mutationResult = result.Append(mutationResult, &mresult)
	}
	// compare (original Resource + patch) vs (original resource)
	// to verify if they are equal
	if !jsonpatch.Equal(patchedResource, rawResource) {
		mresult := result.NewRuleApplicationResult("")
		mresult.FailWithMessagef("Resource does not satisfy the generate policy")
		mutationResult = result.Append(mutationResult, &mresult)
		glog.Info("As objects are different, there is a non applied mutation overlay or patch. So create a violation to inform the user to correct it")
	} else {
		glog.Info("resources are equal, not mutatation required")
	}
	return mutationResult
}
