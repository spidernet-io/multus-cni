// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package spiderpool

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spidernet-io/spiderpool/pkg/constant"
	spider_types "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v2beta1"
	testhelpers "gopkg.in/k8snetworkplumbingwg/multus-cni.v3/pkg/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSpiderpool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "spiderpool")
}

var _ = Describe("Spiderpool", func() {
	Context("Testing multusannotations", func() {
		It("test spiderClaimParameterToAnnotations works", func() {
			scp := &spider_types.SpiderClaimParameter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scp",
					Namespace: "test-ns",
				},
				Spec: spider_types.ClaimParameterSpec{
					StaticNics: []spider_types.StaticNic{{
						MultusConfigName: "nad1",
						Namespace:        "",
					}, {
						MultusConfigName: "nad2",
						Namespace:        "kube-system",
					}},
				},
			}

			pod := testhelpers.NewFakePod("test-pod", "", "")
			Expect(spiderClaimParameterToAnnotations(nil, scp, pod)).ToNot(HaveOccurred())

			Expect(pod.Annotations[constant.MultusDefaultNetAnnot]).To(Equal("default/nad1"))
			Expect(pod.Annotations[constant.MultusNetworkAttachmentAnnot]).To(Equal("kube-system/nad2"))

		})

		It("test spiderClaimParameterToAnnotations: overwrite pod's annotations", func() {
			scp := &spider_types.SpiderClaimParameter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scp",
					Namespace: "test-ns",
				},
				Spec: spider_types.ClaimParameterSpec{
					StaticNics: []spider_types.StaticNic{{
						MultusConfigName: "nad3",
						Namespace:        "kube-system",
					}, {
						MultusConfigName: "nad4",
						Namespace:        "kube-system",
					}},
				},
			}

			pod := testhelpers.NewFakePod("test-pod", "kube-system/nad2", "kube-system/nad1")
			Expect(spiderClaimParameterToAnnotations(nil, scp, pod)).ToNot(HaveOccurred())

			Expect(pod.Annotations[constant.MultusDefaultNetAnnot]).To(Equal("kube-system/nad3"))
			Expect(pod.Annotations[constant.MultusNetworkAttachmentAnnot]).To(Equal("kube-system/nad4"))
		})
	})
})
