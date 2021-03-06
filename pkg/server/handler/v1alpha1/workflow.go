package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// CreateWorkflow ...
func CreateWorkflow(ctx context.Context, project, tenant string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier}
	for _, modifier := range modifiers {
		err := modifier(project, tenant, wf)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Create(wf)
}

// ListWorkflows ...
func ListWorkflows(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	workflows, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get workflows from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflows.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Workflow{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetWorkflow ...
func GetWorkflow(ctx context.Context, project, workflow, tenant string) (*v1alpha1.Workflow, error) {
	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(workflow, metav1.GetOptions{})
}

// UpdateWorkflow ...
func UpdateWorkflow(ctx context.Context, project, workflow, tenant string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(workflow, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWf := origin.DeepCopy()
		newWf.Spec = wf.Spec
		newWf.Annotations = UpdateAnnotations(wf.Annotations, newWf.Annotations)
		_, err = handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Update(newWf)
		return err
	})

	if err != nil {
		return nil, err
	}

	return wf, nil
}

// DeleteWorkflow ...
func DeleteWorkflow(ctx context.Context, project, workflow, tenant string) error {
	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Delete(workflow, nil)
}
