package measurements

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/devices"
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

func buildMeasurements(msg *PipelineMessage, storer MeasurementStorer, archiveTime int) ([]Measurement, error) {
	dev := (*devices.Device)(msg.Device)

	baseMeasurement := Measurement{
		UplinkMessageID:           msg.TracingID,
		OrganisationID:            int(msg.TenantID),
		DeviceID:                  msg.Device.ID,
		DeviceCode:                msg.Device.Code,
		DeviceDescription:         msg.Device.Description,
		DeviceLatitude:            msg.Device.Latitude,
		DeviceLongitude:           msg.Device.Longitude,
		DeviceAltitude:            msg.Device.Altitude,
		DeviceLocationDescription: msg.Device.LocationDescription,
		DeviceProperties:          msg.Device.Properties,
		DeviceState:               msg.Device.State,
		MeasurementLatitude:       msg.Device.Latitude,
		MeasurementLongitude:      msg.Device.Longitude,
		MeasurementAltitude:       msg.Device.Altitude,
		CreatedAt:                 time.Now(),
	}

	measurements := make([]Measurement, len(msg.Measurements))
	for ix, m := range msg.Measurements {

		sensor, err := dev.GetSensorByExternalIDOrFallback(m.SensorExternalID)
		if err != nil {
			return nil, fmt.Errorf("cannot get sensor: %w", err)
		}
		if sensor.ExternalID != m.SensorExternalID {
			m.ObservedProperty = m.SensorExternalID + "_" + m.ObservedProperty
		}

		archiveTimeDays, _ := lo.Coalesce(sensor.ArchiveTime, &archiveTime) // msg.Organisation.ArchiveTime)

		ds, err := storer.GetDatastream(msg.TenantID, sensor.ID, m.ObservedProperty, m.UnitOfMeasurement)
		if err != nil {
			return nil, err
		}

		measurement := baseMeasurement
		measurement.SensorID = sensor.ID
		measurement.SensorCode = sensor.Code
		measurement.SensorDescription = sensor.Description
		measurement.SensorExternalID = sensor.ExternalID
		measurement.SensorProperties = sensor.Properties
		measurement.SensorBrand = sensor.Brand
		measurement.SensorArchiveTime = sensor.ArchiveTime
		measurement.SensorIsFallback = sensor.IsFallback
		measurement.DatastreamID = ds.ID
		measurement.DatastreamDescription = ds.Description
		measurement.DatastreamObservedProperty = ds.ObservedProperty
		measurement.DatastreamUnitOfMeasurement = ds.UnitOfMeasurement
		measurement.MeasurementTimestamp = time.UnixMilli(m.Timestamp)
		measurement.MeasurementValue = m.Value
		measurement.MeasurementProperties = m.Properties
		measurement.MeasurementExpiration = time.UnixMilli(msg.ReceivedAt).Add(time.Duration(*archiveTimeDays) * 24 * time.Hour)

		// Measurement location is either explicitly set or falls back to device location
		if m.Latitude != nil && m.Longitude != nil {
			measurement.MeasurementLatitude = m.Latitude
			measurement.MeasurementLongitude = m.Longitude
			measurement.MeasurementAltitude = m.Altitude
		}

		measurements[ix] = measurement
	}

	return measurements, nil
}
