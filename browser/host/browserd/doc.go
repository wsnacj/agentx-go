// Package browserd provides an explicit, optional host for managing the bundled
// AgentX browser daemon.
//
// The package owns process lifecycle, bundled assets, Playwright bootstrap and
// ownership validation. It does not discover credentials, select a product
// route, or authorize network and process side effects. Callers supply an
// explicit Plan and StatusProbe.
package browserd
