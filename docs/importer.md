# SensorBucket CSV Importer Guide

## Overview

The SensorBucket CSV Importer is a powerful tool that allows you to import, update, and manage your data through simple CSV files. This guide explains how the system works and provides instructions on how to use it effectively.

## Supported Entity Types

The CSV Importer currently supports importing the following types of entities:

1. **Devices and Sensors** - Import devices along with their associated sensors
2. **Features of Interest** - Import geographical features that sensors can be associated with

<figure markdown>
![](../media/importer-devices.png)
<figcaption>The importer screen while importing a CSV of devices and sensors</figcaption>
</figure>

## How It Works

The CSV Importer follows a three-step process:

1. **Upload** - Upload your CSV file containing the entities you want to import
2. **Reconciliation** - The system compares your CSV data with existing data in SensorBucket
3. **Confirmation** - Execute the reconciliation per entity or for all entities in the import

This process ensures that you have full control over what gets created, updated, or deleted in your SensorBucket account.

## CSV File Format

### Devices and Sensors CSV

To import devices and sensors, your CSV file should include the columns below. If the device columns are left empty, but
the sensor columns are filled, the sensor will be associated to the latest device row.

#### Device Columns
- `device code` (required) - Unique identifier for the device
- `device description` - Description of the device
- `device properties X` - Custom properties for the device (replace X with property name)

#### Sensor Columns
- `sensor code` (required) - Unique identifier for the sensor
- `sensor description` - Description of the sensor
- `sensor external_id` - External identifier for the sensor
- `sensor properties X` - Custom properties for the sensor (replace X with property name)

#### Feature of Interest Association
- `sensor feature_of_interest id` - ID of the feature of interest to associate with the sensor
- `sensor feature_of_interest name` - Name of the feature of interest to associate with the sensor
- `sensor feature_of_interest properties X` - Custom properties for the feature of interest

#### Special Columns
- `DELETE` - Add this column and put "DELETE" in the row to mark a device or sensor for deletion

#### Example Devices and Sensors CSV

The sensors below can have a custom attribute named "ultrasonic_correction".

| device code | device description                                      | device latitude | device longitude | device altitude | device properties dev_eui | sensor code | sensor brand     | sensor description         | sensor external_id | sensor feature_of_interest name | sensor properties ultrasonic_correction |
| ----------- | ------------------------------------------------------- | --------------- | ---------------- | --------------- | ------------------------- | ----------- | ---------------- | -------------------------- | ------------------ | ------------------------------- | --------------------------------------- |
| MFM404B     | brug over Havenkanaal, Goes                              | 51.123123       | 3.123123         |                 | D49C10000000404B          | JSN-SR04T   | JSN              | Ultrasonic distance sensor | jsnsr04t           | MPN1115                         | yes                                     |
|             |                                                         |                 |                  |                 |                           | DS18B20     | Maxim Integrated | Temperature sensor         | ds18b20            |                                 |                                         |
|             |                                                         |                 |                  |                 |                           | Antenna     | Lynx WRT         | Antenna                    | antenna            |                                 |                                         |
| MFM4053     | brug over Havenkanaal, Dorp                              | 51.123123      | 3.123123         |                 | D49C100000004053          | JSN-SR04T   | JSN              | Ultrasonic distance sensor | jsnsr04t           | MPN1432                         | yes                                     |
|             |                                                         |                 |                  |                 |                           | DS18B20     | Maxim Integrated | Temperature sensor         | ds18b20            |                                 |                                         |
|             |                                                         |                 |                  |                 |                           | Antenna     | Lynx WRT         | Antenna                    | antenna            |                                 |                                         |

### Features of Interest CSV

To import features of interest, your CSV file should include the following columns:

#### Required Columns
- `name` (required) - Name of the feature of interest

#### Optional Columns
- `description` - Description of the feature of interest
- `properties X` - Custom properties (replace X with property name)
- `latitude` - Latitude coordinate for the feature
- `longitude` - Longitude coordinate for the feature

#### Special Columns
- `DELETE` - Add this column and put "DELETE" in the row to mark a feature for deletion

#### Example Features of Interest CSV
| Name    | Description                                             | Latitude  | Longitude |
| ------- | ------------------------------------------------------- | --------- | --------- |
| MPN1115 | brug over Havenkanaal, nabij Van der Goeskade 8, Goes   | 51.123123 | 3.123123  |
| MPN1432 | brug over Havenkanaal, nabij Brugstraat, Wilhelminadorp | 51.123123 | 3.123123  |


## Import Workflow

### Importing Devices and Sensors

1. Navigate to the Importer page in the SensorBucket web interface
2. Click on "Devices & Sensors"
3. Upload your CSV file
4. The system will analyze your file and compare it with existing data
5. Review the proposed changes:
   - New devices/sensors will be marked for creation
   - Existing devices/sensors will be marked for update
   - Devices/sensors marked with "DELETE" will be scheduled for deletion
6. (Optional) right click a device and select "reconcile this device only" to update a single device.
7. (Optional) click "Reconcile all" to process the entire imported list.

### Importing Features of Interest

1. Navigate to the Importer page in the SensorBucket web interface
2. Click on "Features-of-Interest"
3. Upload your CSV file
4. The system will analyze your file and compare it with existing data
5. Review the proposed changes
6. (Optional) right click a device and select "reconcile this feature-of-interest only" to update a single device.
7. (Optional) click "Reconcile all" to process the entire imported list.

## Tips and Best Practices

1. **Header Detection** - The system automatically detects the header row in your CSV file, but it's best to include it as the first row
2. **Incremental Updates** - You can update just a subset of your devices/sensors by including only those in your CSV
3. **Validation** - The system validates your data before import and will show errors if there are issues
5. **Maintain a single CSV as source** - By maintaining a single CSV per "group" of devices or locations, you will always have a source of truth

## Troubleshooting

### Common Issues

1. **CSV Parse Error** - Ensure your CSV is properly formatted and uses UTF-8 encoding with comma separated values
2. **Missing Required Fields** - Check that all required fields (device code, sensor code, or feature name) are included
3. **Duplicate Codes** - Ensure that device and sensor codes are unique within your CSV
4. **Invalid Coordinates** - For features of interest, ensure latitude and longitude are valid numbers

### Error Messages

- "Could not parse CSV file, is it for devices?" - The CSV format doesn't match the expected format for devices
- "Could not parse CSV file, is it for Features of Interest?" - The CSV format doesn't match the expected format for features
- "Feature of interest must have a name" - The name field is missing for a feature of interest
- "Cannot delete device because Device was not found" - Attempting to delete a device that doesn't exist

## Conclusion

The SensorBucket CSV Importer provides a flexible and powerful way to manage your devices, sensors, and features of interest. By following this guide, you can efficiently import and update your data using simple CSV files.

For additional help or to report issues, please contact the SensorBucket support team.
