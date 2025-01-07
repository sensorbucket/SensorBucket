package measurements

import (
	"context"

	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type PipelineMessage pipeline.Message

func (msg *PipelineMessage) Authorize(keyClient auth.JWKSClient) (context.Context, error) {
	ctx, err := auth.AuthenticateContext(context.Background(), msg.AccessToken, keyClient)
	if err != nil {
		return ctx, err
	}
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_MEASUREMENTS}); err != nil {
		return ctx, err
	}
	msg.TenantID, err = auth.GetTenant(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (msg *PipelineMessage) Validate() error {
	if msg.Device == nil {
		return ErrMissingDeviceInMeasurement
	}
	return nil
}
