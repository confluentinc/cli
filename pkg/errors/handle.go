package errors

func HandleCommon(err error) error {
	if err == nil {
		return nil
	}

	err = catchCCloudTokenErrors(err)
	err = catchCCloudBackendUnmarshallingError(err)
	err = catchTypedErrors(err)
	err = catchMDSErrors(err)
	err = catchCcloudV1Errors(err)
	err = catchOpenAPIError(err)

	return err
}
