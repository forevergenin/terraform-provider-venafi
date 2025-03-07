package venafi

import (
	"fmt"
	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"os"
	"testing"
)

var (
	envSshConfigVariables = fmt.Sprintf(`
variable "TPP_USER" {default = "%s"}
variable "TPP_PASSWORD" {default = "%s"}
variable "TPP_URL" {default = "%s"}
variable "TPP_ZONE" {default = "%s"}
variable "TRUST_BUNDLE" {default = "%s"}
variable "TPP_ACCESS_TOKEN" {default = "%s"}
`,
		os.Getenv("TPP_USER"),
		os.Getenv("TPP_PASSWORD"),
		os.Getenv("TPP_URL"),
		os.Getenv("TPP_ZONE"),
		os.Getenv("TRUST_BUNDLE"),
		os.Getenv("TPP_ACCESS_TOKEN"))

	tokenSshConfigProv = environmentVariables + `
provider "venafi" {
	url = "${var.TPP_URL}"
	access_token = "${var.TPP_ACCESS_TOKEN}"
	trust_bundle = "${file(var.TRUST_BUNDLE)}"
}`

	tokenSshConfigProvWihoutAccessToken = environmentVariables + `
provider "venafi" {
	url = "${var.TPP_URL}"
	trust_bundle = "${file(var.TRUST_BUNDLE)}"
}`

	tppSshConfigResourceTest = `
%s
resource "venafi_ssh_config" "test1" {
	provider = "venafi"
	template="%s"
}
output "ca_public_key"{
	value = venafi_ssh_config.test1.ca_public_key
}
output "principals"{
	value = venafi_ssh_config.test1.principals
}`
)

func TestSshConfig(t *testing.T) {
	t.Log("Testing getting ssh config that returns the CA Public Key and the principals from TPP")

	data := getTestConfigData()

	config := fmt.Sprintf(tppSshConfigResourceTest, tokenSshConfigProv, data.template)
	t.Logf("Testing SSH config with config:\n %s", config)
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					err := checkSshCaPubKey(t, &data, s)
					if err != nil {
						return err
					}
					err = checkSshPrincipals(t, &data, s)
					if err != nil {
						return err
					}
					return nil
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkSshCaPubKey(t *testing.T, data *testData, s *terraform.State) error {
	t.Log("Checking for CA public key", data.template)
	caPubKeyUntyped := s.RootModule().Outputs["ca_public_key"].Value
	caPubKey, ok := caPubKeyUntyped.(string)
	if !ok {
		return fmt.Errorf("output for \"ca_pub_key\" is not a string")
	}
	if caPubKey == "" {
		return fmt.Errorf("The CA public key attribute \"ca_pub_key\" is empty")
	}
	return nil
}

func checkSshPrincipals(t *testing.T, data *testData, s *terraform.State) error {
	t.Log("Checking for principals", data.template)
	principalUntyped := s.RootModule().Outputs["principals"].Value
	err := validateStringListFromSchemaAttribute(principalUntyped, "principals")
	if err != nil {
		return err
	}
	return nil
}

func getTestConfigData() testData {
	return testData{
		template: os.Getenv("TPP_SSH_CA"),
	}
}
