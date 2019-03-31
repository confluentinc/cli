package s3

type S3ObjectTemplate struct {

}
//
//func validateS3Key(key string) (bool, string) {
//	// Format: S3BinPrefix/NAME-VERSION-OS-ARCH
//	split := strings.Split(key, "-")
//
//	// Skip files that don't match our naming standards for binaries
//	if len(split) != 4 {
//		return false, ""
//	}
//
//	// Skip non-matching binaries
//	if split[0] != fmt.Sprintf("%s/%s", r.S3BinPrefix, name) {
//		continue
//	}
//
//	// Skip binaries not for this OS
//	if split[2] != runtime.GOOS {
//		continue
//	}
//
//	// Skip binaries not for this Arch
//	if split[3] != runtime.GOARCH {
//		continue
//	}
//
//}
