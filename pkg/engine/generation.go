package engine

import (
	"github.com/golang/glog"
	kubepolicy "github.com/nirmata/kyverno/pkg/apis/policy/v1alpha1"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/nirmata/kyverno/pkg/result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Generate should be called to process generate rules on the resource
func Generate(client *client.Client, policy kubepolicy.Policy, rawResource []byte, gvk metav1.GroupVersionKind, processExistingResources bool) result.Result {
	policyResult := result.NewPolicyApplicationResult(policy.Name)
	// Generate resource on Namespace creation only
	if gvk.Kind != "Namespace" {
		ruleApplicationResult := result.NewRuleApplicationResult("")
		ruleApplicationResult.FailWithMessagef("Generate is supported for 'Namespace', not applicable to provided kind %s \n", gvk.Kind)
		policyResult = result.Append(policyResult, &ruleApplicationResult)
		return policyResult
	}

	for _, rule := range policy.Spec.Rules {
		if rule.Generation == nil {
			continue
		}
		ruleApplicationResult := result.NewRuleApplicationResult(rule.Name)
		ok := ResourceMeetsDescription(rawResource, rule.ResourceDescription, gvk)

		if !ok {
			glog.Infof("Rule is not applicable to the request: rule name = %s in policy %s \n", rule.Name, policy.ObjectMeta.Name)
			continue
		}

		ruleGeneratorResult := applyRuleGenerator(client, rawResource, rule.Generation, gvk, processExistingResources)
		ruleGeneratorResult.MergeWith(&ruleGeneratorResult)

		policyResult = result.Append(policyResult, &ruleApplicationResult)
	}
	return policyResult
}

func applyRuleGenerator(client *client.Client, rawResource []byte, generator *kubepolicy.Generation, gvk metav1.GroupVersionKind, processExistingResources bool) result.RuleApplicationResult {
	var err error
	ruleGeneratorResult := result.NewRuleApplicationResult("")

	namespace := ParseNameFromObject(rawResource)
	err = client.GenerateResource(*generator, namespace, processExistingResources)
	if err != nil {
		ruleGeneratorResult.FailWithMessagef("Failed to apply generator for %s '%s/%s' : %v", generator.Kind, namespace, generator.Name, err)
	} else {
		ruleGeneratorResult.AddMessagef("Successfully applied generator %s '%s/%s'", generator.Kind, namespace, generator.Name)
	}
	return ruleGeneratorResult
}
