package openapi

import (
	"fmt"
	"os"

	"github.com/afifksupriyadi/crm-handai-backend/cmd/server"
	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/lib/transport"
	"github.com/spf13/cobra"
)

// OpenAPICmd command to print the OpenAPI spec.
var OpenAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "Print the OpenAPI spec",
	Long: `Print the OpenAPI specification for the API in YAML format.
The specification is generated based on the current API routes and their documentation.`,
}

// openAPI generates and prints the OpenAPI specification for the API.
// It takes a cobra command and its arguments as input and returns an error if the operation fails.
// The function will write the OpenAPI spec to stdout.
func openAPI(cmd *cobra.Command, args []string) error {
	c := config.Get()
	f := transport.InitFiber(c)
	api := server.RegisterRoutes(f)

	output, err := api.OpenAPI().YAML()
	if err != nil {
		return fmt.Errorf("failed to generate OpenAPI spec: %w", err)
	}

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

func init() {
	OpenAPICmd.RunE = openAPI
}
