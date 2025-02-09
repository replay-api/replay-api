package iam_query_services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log/slog"
	"math/big"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_dtos "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
)

// WellKnownService is responsible for retrieving well-known configuration data.
type WellKnownService struct {
	Config    common.Config
	JwkReader iam_out.JwkReader
	JwkWriter iam_out.JwkWriter
}

// NewWellKnownService creates a new WellKnownService.
func NewWellKnownService(config common.Config, jwkReader iam_out.JwkReader, jwtWriter iam_out.JwkWriter) *WellKnownService {
	return &WellKnownService{
		Config:    config,
		JwkReader: jwkReader,
		JwkWriter: jwtWriter,
	}
}

// GetOpenConfiguration retrieves the OpenID Connect configuration.
//
// Parameters:
//   - ctx: A context.Context for handling timeouts and cancellations.
//
// Returns:
//   - iam_dtos.OpenConfigurationDTO: A DTO containing the OpenID Connect configuration.
//   - error: Always returns nil as there are no error conditions in the current implementation.
func (s *WellKnownService) GetOpenConfiguration(ctx context.Context) (iam_dtos.OpenConfigurationDTO, error) {
	issuer := s.Config.Auth.Issuer

	slog.Info("Configuration", "config", s.Config)

	openConfiguration := iam_dtos.OpenConfigurationDTO{
		Issuer:                                     issuer,
		AuthorizationEndpoint:                      issuer + "/connect/authorize",
		TokenEndpoint:                              issuer + "/connect/token",
		UserinfoEndpoint:                           issuer + "/connect/userinfo",
		EndSessionEndpoint:                         issuer + "/connect/endsession",
		CheckSessionIframe:                         issuer + "/connect/checksession",
		RevocationEndpoint:                         issuer + "/connect/revocation",
		IntrospectionEndpoint:                      issuer + "/connect/introspect",
		DeviceAuthorizationEndpoint:                issuer + "/connect/deviceauthorization",
		FrontchannelLogoutSupported:                true,
		FrontchannelLogoutSessionSupported:         true,
		BackchannelLogoutSupported:                 true,
		BackchannelLogoutSessionSupported:          true,
		ScopesSupported:                            []string{"openid", "profile", "email", "address", "phone", "offline_access"},
		ClaimsSupported:                            []string{"sub", "name", "family_name", "given_name", "middle_name", "nickname", "preferred_username", "profile", "picture", "website", "gender", "birthdate", "zoneinfo", "locale", "updated_at", "email", "email_verified", "address", "phone_number", "phone_number_verified"},
		GrantTypesSupported:                        []string{"authorization_code", "client_credentials", "refresh_token", "implicit", "password", "urn:ietf:params:oauth:grant-type:device_code"},
		ResponseTypesSupported:                     []string{"code", "token", "id_token", "id_token token", "code id_token", "code token", "code id_token token"},
		ResponseModesSupported:                     []string{"form_post", "query", "fragment"},
		TokenEndpointAuthMethodsSupported:          []string{"client_secret_basic", "client_secret_post"},
		SubjectTypesSupported:                      []string{"public"},
		IdTokenSigningAlgValuesSupported:           []string{"RS256"},
		CodeChallengeMethodsSupported:              []string{"plain", "S256"},
		RequestParameterSupported:                  true,
		RequestObjectSigningAlgValuesSupported:     []string{"RS256", "ES256"},
		AuthorizationResponseIssParameterSupported: true,
		JwksURI: issuer + "/.well-known/openid-configuration/jwks",
	}

	return openConfiguration, nil
}

// GetOpenConfigurationJwks retrieves and prepares the JSON Web Key Set (JWKS) for the OpenID Configuration.
//
// Parameters:
//   - ctx: A context.Context for handling timeouts and cancellations.
//
// Returns:
//   - iam_dtos.OpenConfigurationJwksDTO: A DTO containing the JWKS keys.
//   - error: An error if any step in the process fails, nil otherwise.
func (s *WellKnownService) GetOpenConfigurationJwks(ctx context.Context) (iam_dtos.OpenConfigurationJwksDTO, error) {
	jwks, err := s.JwkReader.Search(ctx, common.Search{})

	if err != nil {
		return iam_dtos.OpenConfigurationJwksDTO{}, err
	}

	jwks, err = s.CreateMissingJwks(ctx, jwks)

	if err != nil {
		return iam_dtos.OpenConfigurationJwksDTO{}, err
	}

	keys := make([]iam_dtos.OpenConfigurationJwksKeyDTO, len(jwks))

	for i, jwk := range jwks {
		keys[i] = iam_dtos.OpenConfigurationJwksKeyDTO{
			Kid: jwk.Kid,
			Kty: jwk.Kty,
			E:   jwk.E,
			N:   jwk.N,
			Use: jwk.Use,
			Alg: jwk.Alg,
		}
	}

	result := iam_dtos.OpenConfigurationJwksDTO{
		Keys: keys,
	}

	return result, nil
}

// CreateMissingJwks generates and adds new JSON Web Keys (JWKs) to ensure a minimum of 10 keys are available.
//
// Parameters:
//   - ctx: A context.Context for handling timeouts and cancellations.
//   - keys: A slice of existing iam_entities.Jwk representing the current set of JWKs.
//
// Returns:
//   - []iam_entities.Jwk: An updated slice of JWKs, including any newly created keys.
//   - error: An error if key generation, conversion, or storage fails; nil otherwise.
func (s *WellKnownService) CreateMissingJwks(ctx context.Context, keys []iam_entities.Jwk) ([]iam_entities.Jwk, error) {
	if len(keys) < 10 {
		for i := len(keys); i < 10; i++ {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			publicKey := privateKey.Public().(*rsa.PublicKey)

			if err != nil {
				return nil, err
			}

			privateKeyPEM := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}
			privateKeyPEMBytes := pem.EncodeToMemory(privateKeyPEM)

			privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyPEMBytes)

			newKey := iam_entities.Jwk{
				Kid:        uuid.New(),
				Kty:        "RSA",
				E:          base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
				N:          base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				Use:        "sig",
				Alg:        "RS256",
				PrivateKey: privateKeyBase64,
			}

			_, err = s.JwkWriter.Create(ctx, &newKey)

			if err != nil {
				return nil, err
			}

			keys = append(keys, newKey)
		}
	}

	return keys, nil
}
