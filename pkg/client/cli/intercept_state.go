package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	manager "github.com/datawire/telepresence2/pkg/rpc"
	"github.com/datawire/telepresence2/pkg/rpc/connector"
)

type interceptState struct {
	cmd *cobra.Command
	cc  connector.ConnectorClient
	ir  *manager.CreateInterceptRequest
}

func newInterceptState(cs connector.ConnectorClient, ir *manager.CreateInterceptRequest, cmd *cobra.Command) *interceptState {
	return &interceptState{cc: cs, ir: ir, cmd: cmd}
}

func (is *interceptState) EnsureState() (bool, error) {
	r, err := is.cc.AddIntercept(context.Background(), is.ir)
	if err != nil {
		return false, err
	}
	switch r.Error {
	case connector.InterceptError_UNSPECIFIED:
		fmt.Fprintf(is.cmd.OutOrStdout(), "Using deployment %s in namespace %s\n", is.ir.InterceptSpec.Name, r.ErrorText)

		return true, nil
	case connector.InterceptError_ALREADY_EXISTS:
		fmt.Fprintln(is.cmd.OutOrStdout(), interceptMessage(r.Error, r.ErrorText))
		return false, nil
	case connector.InterceptError_NO_CONNECTION:
		return false, connectorIsNotRunning
	}
	return false, errors.New(interceptMessage(r.Error, r.ErrorText))
}

func (is *interceptState) DeactivateState() error {
	name := strings.TrimSpace(is.ir.InterceptSpec.Name)
	var r *connector.InterceptResult
	var err error
	r, err = is.cc.RemoveIntercept(context.Background(), &connector.RemoveInterceptRequest{Name: name})
	if err != nil {
		return err
	}
	if r.Error != connector.InterceptError_UNSPECIFIED {
		return errors.New(interceptMessage(r.Error, r.ErrorText))
	}
	return nil
}
