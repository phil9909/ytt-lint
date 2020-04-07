package pull

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func Pull() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("user.Current(): %w", err)
	}
	schemaDir := path.Join(usr.HomeDir, ".ytt-lint", "schema", "k8s")

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("kubeConfig.ClientConfig(): %w", err)
	}

	crdClientset, err := apiextensionsclientset.NewForConfig(config)

	crds, err := crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(): %w", err)
		}
		fmt.Println("Error for ApiextensionsV1. Falling back to ApiextensionsV1beta1.")
		crds, err := crdClientset.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(): %w", err)
		}

		metadataTemplate := apiextensionsv1beta1.JSONSchemaProps{}
		err = json.Unmarshal([]byte(metadataTemplateJSON), &metadataTemplate)
		if err != nil {
			return fmt.Errorf("json.Unmarshal([]byte(metadataTemplateJSON), &metadataTemplate): %w", err)
		}

		for _, crd := range crds.Items {
			kind := crd.Spec.Names.Kind
			group := crd.Spec.Group

			for _, version := range crd.Spec.Versions {
				var schema *apiextensionsv1beta1.JSONSchemaProps
				if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
					if crd.Spec.Validation == nil || crd.Spec.Validation.OpenAPIV3Schema == nil {
						fmt.Printf("%s version %s of group %s does not contain a schema and will be skipped.\n", kind, version.Name, group)
						continue
					}
					schema = crd.Spec.Validation.OpenAPIV3Schema
				} else {
					schema = version.Schema.OpenAPIV3Schema
				}
				dirname := path.Join(schemaDir, group, version.Name)
				filename := path.Join(dirname, strings.ToLower(kind)+".json")

				schema.Properties["metadata"] = metadataTemplate
				if _, ok := schema.Properties["kind"]; !ok {
					schema.Properties["kind"] = apiextensionsv1beta1.JSONSchemaProps{
						Type: "string",
					}
				}
				if _, ok := schema.Properties["apiVersion"]; !ok {
					schema.Properties["apiVersion"] = apiextensionsv1beta1.JSONSchemaProps{
						Type: "string",
					}
				}

				err = os.MkdirAll(dirname, os.ModePerm)
				if err != nil {
					return fmt.Errorf("os.MkdirAll: %w", err)
				}

				data, err := json.Marshal(schema)
				if err != nil {
					return fmt.Errorf("json.Marshal(schema): %w", err)
				}

				fmt.Printf("Writing schema for %s version %s of group %s to %s\n", kind, version.Name, group, filename)
				err = ioutil.WriteFile(filename, data, os.ModePerm)
				if err != nil {
					return fmt.Errorf("ioutil.WriteFile: %w", err)
				}
			}
		}

	} else {
		metadataTemplate := apiextensionsv1.JSONSchemaProps{}
		err = json.Unmarshal([]byte(metadataTemplateJSON), &metadataTemplate)
		if err != nil {
			return fmt.Errorf("json.Unmarshal([]byte(metadataTemplateJSON), &metadataTemplate): %w", err)
		}

		for _, crd := range crds.Items {
			kind := crd.Spec.Names.Kind
			group := crd.Spec.Group

			for _, version := range crd.Spec.Versions {
				var schema *apiextensionsv1.JSONSchemaProps
				if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
					continue
				} else {
					schema = version.Schema.OpenAPIV3Schema
				}
				dirname := path.Join(schemaDir, group, version.Name)
				filename := path.Join(dirname, strings.ToLower(kind)+".json")

				schema.Properties["metadata"] = metadataTemplate
				if _, ok := schema.Properties["kind"]; !ok {
					schema.Properties["kind"] = apiextensionsv1.JSONSchemaProps{
						Type: "string",
					}
				}
				if _, ok := schema.Properties["apiVersion"]; !ok {
					schema.Properties["apiVersion"] = apiextensionsv1.JSONSchemaProps{
						Type: "string",
					}
				}

				err = os.MkdirAll(dirname, os.ModePerm)
				if err != nil {
					return fmt.Errorf("os.MkdirAll: %w", err)
				}

				data, err := json.Marshal(schema)
				if err != nil {
					return fmt.Errorf("json.Marshal(schema): %w", err)
				}

				fmt.Printf("Writing schema for %s version %s of group %s to %s\n", kind, version.Name, group, filename)
				err = ioutil.WriteFile(filename, data, os.ModePerm)
				if err != nil {
					return fmt.Errorf("ioutil.WriteFile: %w", err)
				}
			}
		}
	}

	return nil
}

const metadataTemplateJSON = `
{
	"description": "ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.",
	"type": "object",
	"properties": {
		"annotations": {
			"description": "Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations",
			"type": "object",
			"additionalProperties": {
				"type": "string"
			}
		},
		"clusterName": {
			"description": "The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.",
			"type": "string"
		},
		"creationTimestamp": {
			"description": "Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers.",
			"type": "string",
			"format": "date-time"
		},
		"deletionGracePeriodSeconds": {
			"description": "Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only.",
			"type": "integer",
			"format": "int64"
		},
		"deletionTimestamp": {
			"description": "Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers.",
			"type": "string",
			"format": "date-time"
		},
		"finalizers": {
			"description": "Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list.",
			"type": "array",
			"items": {
				"type": "string"
			}
		},
		"generateName": {
			"description": "GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.\n\nIf this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header).\n\nApplied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency",
			"type": "string"
		},
		"generation": {
			"description": "A sequence number representing a specific generation of the desired state. Populated by the system. Read-only.",
			"type": "integer",
			"format": "int64"
		},
		"labels": {
			"description": "Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
			"type": "object",
			"additionalProperties": {
				"type": "string"
			}
		},
		"managedFields": {
			"description": "ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like \"ci-cd\". The set of fields is always in the version that the workflow used when modifying the object.",
			"type": "array",
			"items": {
				"description": "ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource that the fieldset applies to.",
				"type": "object",
				"properties": {
					"apiVersion": {
						"description": "APIVersion defines the version of this resource that this field set applies to. The format is \"group/version\" just like the top-level APIVersion field. It is necessary to track the version of a field set because it cannot be automatically converted.",
						"type": "string"
					},
					"fieldsType": {
						"description": "FieldsType is the discriminator for the different fields format and version. There is currently only one possible value: \"FieldsV1\"",
						"type": "string"
					},
					"fieldsV1": {
						"description": "FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.\n\nEach key is either a '.' representing the field itself, and will always map to an empty set, or a string representing a sub-field or item. The string will follow one of these four formats: 'f:\u003cname\u003e', where \u003cname\u003e is the name of a field in a struct, or key in a map 'v:\u003cvalue\u003e', where \u003cvalue\u003e is the exact json formatted value of a list item 'i:\u003cindex\u003e', where \u003cindex\u003e is position of a item in a list 'k:\u003ckeys\u003e', where \u003ckeys\u003e is a map of  a list item's key fields to their unique values If a key maps to an empty Fields value, the field that key represents is part of the set.\n\nThe exact format is defined in sigs.k8s.io/structured-merge-diff",
						"type": "object"
					},
					"manager": {
						"description": "Manager is an identifier of the workflow managing these fields.",
						"type": "string"
					},
					"operation": {
						"description": "Operation is the type of operation which lead to this ManagedFieldsEntry being created. The only valid values for this field are 'Apply' and 'Update'.",
						"type": "string"
					},
					"time": {
						"description": "Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers.",
						"type": "string",
						"format": "date-time"
					}
				}
			}
		},
		"name": {
			"description": "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			"type": "string"
		},
		"namespace": {
			"description": "Namespace defines the space within each name must be unique. An empty namespace is equivalent to the \"default\" namespace, but \"default\" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.\n\nMust be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces",
			"type": "string"
		},
		"ownerReferences": {
			"description": "List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller.",
			"type": "array",
			"items": {
				"description": "OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field.",
				"type": "object",
				"required": [
					"apiVersion",
					"kind",
					"name",
					"uid"
				],
				"properties": {
					"apiVersion": {
						"description": "API version of the referent.",
						"type": "string"
					},
					"blockOwnerDeletion": {
						"description": "If true, AND if the owner has the \"foregroundDeletion\" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. Defaults to false. To set this field, a user needs \"delete\" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned.",
						"type": "boolean"
					},
					"controller": {
						"description": "If true, this reference points to the managing controller.",
						"type": "boolean"
					},
					"kind": {
						"description": "Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
						"type": "string"
					},
					"name": {
						"description": "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
						"type": "string"
					},
					"uid": {
						"description": "UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids",
						"type": "string"
					}
				}
			}
		},
		"resourceVersion": {
			"description": "An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.\n\nPopulated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency",
			"type": "string"
		},
		"selfLink": {
			"description": "SelfLink is a URL representing this object. Populated by the system. Read-only.\n\nDEPRECATED Kubernetes will stop propagating this field in 1.20 release and the field is planned to be removed in 1.21 release.",
			"type": "string"
		},
		"uid": {
			"description": "UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.\n\nPopulated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids",
			"type": "string"
		}
	}
}
`
