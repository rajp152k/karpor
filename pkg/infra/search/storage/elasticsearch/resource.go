// Copyright The Karpor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KusionStack/karpor/pkg/infra/search/storage"
	"github.com/elliotxx/esquery"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	resourceKeyCluster           = "cluster"
	resourceKeyAPIVersion        = "apiVersion"
	resourceKeyKind              = "kind"
	resourceKeyNamespace         = "namespace"
	resourceKeyName              = "name"
	resourceKeyLabels            = "labels"
	resourceKeyAnnotations       = "annotations"
	resourceKeyCreationTimestamp = "creationTimestamp"
	resourceKeyDeletionTimestamp = "deletionTimestamp"
	resourceKeyOwnerReferences   = "ownerReferences"
	resourceKeyResourceVersion   = "resourceVersion"
	resourceKeyContent           = "content"
	resourceKeySyncAt            = "syncAt"  // resource save/update/delete time
	resourceKeyDeleted           = "deleted" // indicates whether the resource is deleted in cluster
)

var ErrNotFound = fmt.Errorf("object not found")

// SaveResource stores an object in the Elasticsearch storage for the specified cluster.
func (s *Storage) SaveResource(ctx context.Context, cluster string, obj runtime.Object) error {
	id, body, err := s.generateResourceDocument(cluster, obj)
	if err != nil {
		return err
	}
	return s.client.SaveDocument(ctx, s.resourceIndexName, id, bytes.NewReader(body))
}

// Refresh will update ES index. If you want the previous document changes to be
// searchable immediately, you need to call refresh manually.
//
// Refer to https://www.elastic.co/guide/en/elasticsearch/guide/current/near-real-time.html to see detail.
func (s *Storage) Refresh(ctx context.Context) error {
	return s.client.Refresh(ctx, s.resourceIndexName)
}

// SoftDeleteResource only sets the deleted field to true, not really deletes the data in storage.
func (s *Storage) SoftDeleteResource(ctx context.Context, cluster string, obj runtime.Object) error {
	unObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		// TODO: support other implement of runtime.Object
		return fmt.Errorf("only support *unstructured.Unstructured type")
	}

	if err := s.GetResource(ctx, cluster, unObj); err != nil {
		return err
	}

	body, err := json.Marshal(map[string]map[string]interface{}{
		"doc": {
			resourceKeySyncAt:  time.Now(),
			resourceKeyDeleted: true,
		},
	})
	if err != nil {
		return err
	}

	id := string(unObj.GetUID())
	return s.client.UpdateDocument(ctx, s.resourceIndexName, id, bytes.NewReader(body))
}

// DeleteResource removes an object from the Elasticsearch storage for the specified cluster.
func (s *Storage) DeleteResource(ctx context.Context, cluster string, obj runtime.Object) error {
	unObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		// TODO: support other implement of runtime.Object
		return fmt.Errorf("only support *unstructured.Unstructured type")
	}

	if err := s.GetResource(ctx, cluster, unObj); err != nil {
		return err
	}

	return s.client.DeleteDocument(ctx, s.resourceIndexName, string(unObj.GetUID()))
}

// GetResource retrieves an object from the Elasticsearch storage for the specified cluster.
func (s *Storage) GetResource(ctx context.Context, cluster string, obj runtime.Object) error {
	unObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		// TODO: support other implement of runtime.Object
		return fmt.Errorf("only support *unstructured.Unstructured type")
	}

	query := generateResourceQuery(cluster, unObj.GetNamespace(), unObj.GetName(), unObj)
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(query); err != nil {
		return err
	}
	resp, err := s.client.SearchDocument(ctx, s.resourceIndexName, buf)
	if err != nil {
		return err
	}

	if len(resp.Hits.Hits) == 0 {
		return fmt.Errorf("no resource found for cluster: %s, namespace: %s, name: %s", cluster, unObj.GetNamespace(), unObj.GetName())
	}

	res, err := storage.Map2Resource(resp.Hits.Hits[0].Source)
	if err != nil {
		return err
	}

	unObj.Object = res.Object
	return nil
}

// CountResources return a count of resources in the Elasticsearch storage.
func (s *Storage) CountResources(ctx context.Context) (int, error) {
	if resp, err := s.client.Count(ctx, s.resourceIndexName); err != nil {
		return 0, err
	} else {
		return int(resp.Count), nil
	}
}

// DeleteAllResources removes all resources from the Elasticsearch storage for the specified cluster.
func (s *Storage) DeleteAllResources(ctx context.Context, cluster string) error {
	query := make(map[string]interface{})
	query["query"] = esquery.Bool().Must(
		esquery.Term(resourceKeyCluster, cluster),
	).Map()
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(query); err != nil {
		return err
	}
	return s.client.DeleteDocumentByQuery(ctx, s.resourceIndexName, buf)
}

// generateResourceDocument creates an resource document for Elasticsearch with
// the specified cluster and object.
func (s *Storage) generateResourceDocument(cluster string, obj runtime.Object) (id string, body []byte, err error) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer([]byte{})
	if err = s.objectEncoder.Encode(obj, buf); err != nil {
		return
	}

	body, err = json.Marshal(map[string]interface{}{
		resourceKeyCluster:           cluster,
		resourceKeyAPIVersion:        obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		resourceKeyKind:              obj.GetObjectKind().GroupVersionKind().Kind,
		resourceKeyNamespace:         metaObj.GetNamespace(),
		resourceKeyName:              metaObj.GetName(),
		resourceKeyLabels:            metaObj.GetLabels(),
		resourceKeyAnnotations:       metaObj.GetAnnotations(),
		resourceKeyCreationTimestamp: metaObj.GetCreationTimestamp(),
		resourceKeyDeletionTimestamp: metaObj.GetDeletionTimestamp(),
		resourceKeyOwnerReferences:   metaObj.GetOwnerReferences(),
		resourceKeyResourceVersion:   metaObj.GetResourceVersion(),
		resourceKeyContent:           buf.String(),
		resourceKeySyncAt:            time.Now(),
		resourceKeyDeleted:           false,
	})
	if err != nil {
		return
	}
	id = string(metaObj.GetUID())
	return
}

// generateResourceQuery creates a query to search for an object in
// Elasticsearch based on resource's cluster, namespace, and name.
func generateResourceQuery(cluster, namespace, name string, obj runtime.Object) map[string]interface{} {
	query := make(map[string]interface{})
	query["query"] = esquery.Bool().Must(
		esquery.Term(resourceKeyAPIVersion, obj.GetObjectKind().GroupVersionKind().GroupVersion().String()),
		esquery.Term(resourceKeyKind, obj.GetObjectKind().GroupVersionKind().Kind),
		esquery.Term(resourceKeyName, name),
		esquery.Term(resourceKeyNamespace, namespace),
		esquery.Term(resourceKeyCluster, cluster),
	).Map()
	return query
}

// CheckStorageHealth checks the health of the Elasticsearch storage by pinging the client.
func (s *Storage) CheckStorageHealth(ctx context.Context) error {
	return s.client.CheckElasticSearchLiveness(ctx)
}
