package dcos

import (
	"context"
	"fmt"
	"log"

	"github.com/dcos/client-go/dcos"
)

func iamErrorOK(err error) (dcos.IamError, bool) {
	if err != nil {
		// is dcos.GenericOpenAPIError
		if apiErr, ok := err.(dcos.GenericOpenAPIError); ok {
			if iamErr, ok := apiErr.Model().(dcos.IamError); ok {
				return iamErr, true
			}
			log.Printf("[WARNING] iamErrorOK GenericOpenAPIError - %v", apiErr)
		}
	}
	return dcos.IamError{}, false
}

func iamEnsureRid(ctx context.Context, client *dcos.APIClient, rid string) error {
	ridRes, resp, err := client.IAM.GetResourceACLs(ctx, rid)
	log.Printf("[TRACE] EnsureRID GetResourceACLs - %v", resp)

	if iamErr, ok := iamErrorOK(err); ok {
		switch iamErr.Code {
		case dcos.IAM_ERR_UNKNOWN_RESOURCE_ID:
			log.Printf("[TRACE] EnsureRID GetResourceACLs Rid: %s not found. Will be created", rid)
			_, err = client.IAM.CreateResourceACL(ctx, rid, dcos.IamaclCreate{})
			return err
		case dcos.IAM_ERR_INVALID_RESOURCE_ID:
			log.Printf("[ERROR] Invalid Resource ID %s was provided", rid)
			return err
		default:
			return err
		}
	}

	if err != nil {
		// Unexpect error returned.
		log.Printf("[CRITICAL] ensureRid unexpected error for rid: %s returned: %v", rid, err)
		return err
	}

	if ridRes.Rid == "" {
		return fmt.Errorf("Found %s but return with empty .Rid %v", rid, ridRes)
	}

	return nil
}

func iamPrettyError(iamErr dcos.IamError) error {
	return fmt.Errorf("IamError %s - %s - %s", iamErr.Title, iamErr.Code, iamErr.Description)
}
