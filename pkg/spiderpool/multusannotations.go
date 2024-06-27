// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package spiderpool

import (
	"fmt"
	"strings"

	"github.com/spidernet-io/spiderpool/pkg/constant"
	spider_types "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v2beta1"
	k8s "gopkg.in/k8snetworkplumbingwg/multus-cni.v3/pkg/k8sclient"
	"gopkg.in/k8snetworkplumbingwg/multus-cni.v3/pkg/logging"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConvertResourceClaimToAnnotations(kubeClient *k8s.ClientInfo, pod *v1.Pod) error {
	if kubeClient == nil {
		return nil
	}

	// we should check the DRA feature-gate if is enabled in your cluster
	draEnabled, err := kubeClient.IsDynamicResourceAllocationEnabled()
	if err != nil {
		return fmt.Errorf("error checking the DynamicResourceAllocation feature if is enabled: %w", err)
	}

	// if not, we should return directly
	if !draEnabled {
		return nil
	}

	for _, rc := range pod.Spec.ResourceClaims {
		if rc.Source.ResourceClaimTemplateName != nil {
			rct, err := kubeClient.GetResourceTemplate(*rc.Source.ResourceClaimTemplateName, pod.Namespace)
			if err != nil {
				return fmt.Errorf("failed to get resourceTemplate for pod %s/%s: %v", pod.Namespace, pod.Name, err)
			}

			if rct.Spec.Spec.ResourceClassName == constant.DRADriverName && rct.Spec.Spec.ParametersRef.APIGroup == constant.SpiderpoolAPIGroup &&
				rct.Spec.Spec.ParametersRef.Kind == constant.KindSpiderClaimParameter {
				spc, err := kubeClient.GetSpiderClaimParameter(rct.Spec.Spec.ParametersRef.Name, pod.Namespace)
				if err != nil {
					return fmt.Errorf("failed to get spiderClaimParameter for pod %s/%s: %v", pod.Namespace, pod.Name, err)
				}
				return spiderClaimParameterToAnnotations(kubeClient, spc, pod)
			}
		}
	}
	return nil
}
func spiderClaimParameterToAnnotations(kubeClient *k8s.ClientInfo, scp *spider_types.SpiderClaimParameter, pod *v1.Pod) error {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	} else {
		logging.Debugf("Warning! pod's multus annotations: %v will be overwrite by spidercliamparameter %s/%s", pod.Annotations, scp.Namespace, scp.Name)
		if _, ok := pod.Annotations[constant.MultusDefaultNetAnnot]; ok {
			delete(pod.Annotations, constant.MultusDefaultNetAnnot)
		}

		if _, ok := pod.Annotations[constant.MultusNetworkAttachmentAnnot]; ok {
			delete(pod.Annotations, constant.MultusNetworkAttachmentAnnot)
		}
	}

	for idx, nic := range scp.Spec.StaticNics {
		if kubeClient != nil {
			_, err := kubeClient.GetSpiderMultusConfig(nic.MultusConfigName, nic.Namespace)
			if err != nil {
				return fmt.Errorf("get spiderMultusConfig %s/%s failed: %v", nic.MultusConfigName, nic.Namespace, err)
			}
		}

		if nic.Namespace == "" {
			nic.Namespace = metav1.NamespaceDefault
		}

		value := fmt.Sprintf("%s/%s", nic.Namespace, nic.MultusConfigName)
		if idx == 0 {
			pod.Annotations[constant.MultusDefaultNetAnnot] = value
			continue
		}

		if v, ok := pod.Annotations[constant.MultusNetworkAttachmentAnnot]; ok {
			pod.Annotations[constant.MultusNetworkAttachmentAnnot] = fmt.Sprintf("%v,%s", v, value)
		} else {
			pod.Annotations[constant.MultusNetworkAttachmentAnnot] = value
		}
	}
	return nil
}
func splitStaticNic(nic string) (ns, name string) {
	res := strings.Split(nic, "/")
	if len(res) == 1 {
		return metav1.NamespaceDefault, res[0]
	}

	return res[0], res[1]
}
