package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc defines the evidence module's codec. The codec is not sealed as to
// allow other modules to register their concrete Evidence types.
var ModuleCdc = codec.New()

// RegisterCodec registers all the necessary types and interfaces for the
// evidence module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Evidence)(nil), nil)

	// TODO: Register concrete evidence types.
}

// RegisterEvidenceTypeCodec registers an external concrete Evidence type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitEvidence
// to be correctly Amino encoded and decoded.
func RegisterEvidenceTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

func init() {
	RegisterCodec(ModuleCdc)
}
