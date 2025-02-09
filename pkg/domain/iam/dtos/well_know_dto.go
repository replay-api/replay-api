package iam_dtos

import "github.com/google/uuid"

// OpenConfigurationDTO represents the OpenID Connect (OIDC) configuration.
type OpenConfigurationDTO struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	UserinfoEndpoint                           string   `json:"userinfo_endpoint"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint"`
	CheckSessionIframe                         string   `json:"check_session_iframe"`
	RevocationEndpoint                         string   `json:"revocation_endpoint"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint"`
	DeviceAuthorizationEndpoint                string   `json:"device_authorization_endpoint"`
	JwksURI                                    string   `json:"jwks_uri"`
	FrontchannelLogoutSupported                bool     `json:"frontchannel_logout_supported"`
	FrontchannelLogoutSessionSupported         bool     `json:"frontchannel_logout_session_supported"`
	BackchannelLogoutSupported                 bool     `json:"backchannel_logout_supported"`
	BackchannelLogoutSessionSupported          bool     `json:"backchannel_logout_session_supported"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	RequestParameterSupported                  bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported     []string `json:"request_object_signing_alg_values_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`
}

// OpenConfigurationJwksDTO represents a JSON Web Key Set (JWKS) for the OpenID Connect (OIDC) configuration.
type OpenConfigurationJwksDTO struct {
	Keys []OpenConfigurationJwksKeyDTO `json:"keys"`
}

// OpenConfigurationJwksKeyDTO represents a JSON Web Key (JWK) for the OpenID Connect (OIDC) configuration.
type OpenConfigurationJwksKeyDTO struct {
	Kid uuid.UUID `json:"kid"`
	Kty string    `json:"kty"`
	E   string    `json:"e"`
	N   string    `json:"n"`
	Use string    `json:"use"`
	Alg string    `json:"alg"`
}
