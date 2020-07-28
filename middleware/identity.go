package middleware

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/go-ns/identity"
)

// IdentityHandler is a wrapper around go-ns identity handler, using the provided healthclient to generate an identity client
var IdentityHandler = func(cli *health.Client, url string) func(http.Handler) http.Handler {
	idClient := clientsidentity.NewAPIClient(cli.Client, url)
	return identity.HandlerForHTTPClient(idClient)
}
