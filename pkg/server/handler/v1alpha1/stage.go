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

// CreateStage ...
func CreateStage(ctx context.Context, project, tenant string, stg *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier}
	for _, modifier := range modifiers {
		err := modifier(project, tenant, stg)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Create(stg)
}

// ListStages ...
func ListStages(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	stages, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get stages from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := stages.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetStage ...
func GetStage(ctx context.Context, project, stage, tenant string) (*v1alpha1.Stage, error) {
	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(stage, metav1.GetOptions{})
}

// UpdateStage ...
func UpdateStage(ctx context.Context, project, stage, tenant string, stg *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(stage, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newStg := origin.DeepCopy()
		newStg.Spec = stg.Spec
		newStg.Annotations = UpdateAnnotations(stg.Annotations, newStg.Annotations)
		_, err = handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Update(newStg)
		return err
	})

	if err != nil {
		return nil, err
	}

	return stg, nil
}

// DeleteStage ...
func DeleteStage(ctx context.Context, project, stage, tenant string) error {
	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Delete(stage, nil)
}
