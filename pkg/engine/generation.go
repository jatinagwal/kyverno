package engine

import (
	"time"

	kyvernov1beta1 "github.com/kyverno/kyverno/api/kyverno/v1beta1"
	"github.com/kyverno/kyverno/pkg/autogen"
	engineapi "github.com/kyverno/kyverno/pkg/engine/api"
	"github.com/kyverno/kyverno/pkg/logging"
	"k8s.io/client-go/tools/cache"
)

// GenerateResponse checks for validity of generate rule on the resource
func GenerateResponse(
	contextLoader ContextLoaderFactory,
	policyContext engineapi.PolicyContext,
	gr kyvernov1beta1.UpdateRequest,
) (resp *engineapi.EngineResponse) {
	policyStartTime := time.Now()
	return filterGenerateRules(contextLoader, policyContext, gr.Spec.Policy, policyStartTime)
}

func filterGenerateRules(
	contextLoader ContextLoaderFactory,
	policyContext engineapi.PolicyContext,
	policyNameKey string,
	startTime time.Time,
) *engineapi.EngineResponse {
	newResource := policyContext.NewResource()
	kind := newResource.GetKind()
	name := newResource.GetName()
	namespace := newResource.GetNamespace()
	apiVersion := newResource.GetAPIVersion()
	pNamespace, pName, err := cache.SplitMetaNamespaceKey(policyNameKey)
	if err != nil {
		logging.Error(err, "failed to spilt name and namespace", policyNameKey)
	}
	resp := &engineapi.EngineResponse{
		PolicyResponse: engineapi.PolicyResponse{
			Policy: engineapi.PolicySpec{
				Name:      pName,
				Namespace: pNamespace,
			},
			PolicyStats: engineapi.PolicyStats{
				ExecutionStats: engineapi.ExecutionStats{
					Timestamp: startTime.Unix(),
				},
			},
			Resource: engineapi.ResourceSpec{
				Kind:       kind,
				Name:       name,
				Namespace:  namespace,
				APIVersion: apiVersion,
			},
		},
	}
	if policyContext.ExcludeResourceFunc()(kind, namespace, name) {
		logging.WithName("Generate").Info("resource excluded", "kind", kind, "namespace", namespace, "name", name)
		return resp
	}

	for _, rule := range autogen.ComputeRules(policyContext.Policy()) {
		if ruleResp := filterRule(contextLoader, rule, policyContext); ruleResp != nil {
			resp.PolicyResponse.Rules = append(resp.PolicyResponse.Rules, *ruleResp)
		}
	}

	return resp
}
