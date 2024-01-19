package txnbuild

import (
	"testing"

	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/gxdr"
	"github.com/hcnet/go/randxdr"
	"github.com/hcnet/go/xdr"
	goxdr "github.com/xdrpp/goxdr/xdr"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccountFromXDR(t *testing.T) {
	txeB64 := "AAAAAMOrP0B2tL9IUn5QL8nn8q88kkFui1x3oW9omCj6hLhfAAAAZAAAAMcAAAAWAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAAAAAAAAAAJ5yfHhgKAxylgecjAymWqNzLWRk/MqSYt+X9duZ2DfyAAAAF0h26AAAAAAAAAAAAvqEuF8AAABAZ5q2N2BHRylT28T1DbUVU7QKTbKZ+6DLefzJoCjHo2T8vcI/PjF8gsRu/r2M60Uzcw3WmqRFerA6DnJILIEdDoZW4JwAAABAsFL3WXr+tDK5tjR/0ZBVuNyzyqSa8Li2tUMUmB23PWuPG71ObUPTShkhlc7ydNN/qYRaA/Mafm+vsIQWDbCRDA=="

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var ca CreateAccount
		err = ca.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "GAIH3ULLFQ4DGSECF2AR555KZ4KNDGEKN4AFI4SU2M7B43MGK3QJZNSR", ca.SourceAccount, "source accounts should match")
			assert.Equal(t, "GCPHE7DYMAUAY4UWA6OIYDFGLKRXGLLEMT6MVETC36L7LW4Z3A37EJW5", ca.Destination, "destination should match")
			assert.Equal(t, "10000.0000000", ca.Amount, "starting balance should match")
		}
	}

	txeB64NoSource := "AAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAAAyAAADXYAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAAAAAdgRnAAAAAAAAAAAA"
	xdrEnv, err = unmarshalBase64(txeB64NoSource)
	if assert.NoError(t, err) {
		var ca CreateAccount
		err = ca.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "", ca.SourceAccount, "source accounts should match")
			assert.Equal(t, "GAIH3ULLFQ4DGSECF2AR555KZ4KNDGEKN4AFI4SU2M7B43MGK3QJZNSR", ca.Destination, "destination should match")
			assert.Equal(t, "198.0000000", ca.Amount, "starting balance should match")
		}
	}
}

func TestZeroBalanceAccount(t *testing.T) {
	sponsor, sponsee := newKeypair0(), newKeypair1()
	ops := []Operation{
		&BeginSponsoringFutureReserves{SponsoredID: sponsee.Address()},
		&CreateAccount{
			Destination: sponsee.Address(),
			Amount:      "0",
		},
		&EndSponsoringFutureReserves{
			SourceAccount: sponsee.Address(),
		},
	}

	sourceAccount := SimpleAccount{AccountID: sponsor.Address()}
	_, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			Operations:           ops,
			IncrementSequenceNum: true,
			BaseFee:              MinBaseFee,
			Preconditions:        Preconditions{TimeBounds: NewInfiniteTimeout()},
		},
	)

	assert.NoErrorf(t, err, "zero-balance account creation should work")
}

func TestPaymentFromXDR(t *testing.T) {
	txeB64 := "AAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAABkAAADXYAAAABAAAAAAAAAAAAAAACAAAAAQAAAABooIjdquKgVwBESJjRJ3KWgA8pcgUbc3aqRLJLugjdMQAAAAEAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAAAAAAAAAAAX14QAAAAAAAAAAAQAAAAAQfdFrLDgzSIIugR73qs8U0ZiKbwBUclTTPh5thlbgnAAAAAFYWQAAAAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAAAAE/exwAAAAAAAAAAAA=="

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var p Payment
		err = p.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", p.SourceAccount, "source accounts should match")
			assert.Equal(t, "GAIH3ULLFQ4DGSECF2AR555KZ4KNDGEKN4AFI4SU2M7B43MGK3QJZNSR", p.Destination, "destination should match")
			assert.Equal(t, "10.0000000", p.Amount, "amount should match")
			assert.Equal(t, true, p.Asset.IsNative(), "Asset should be native")
		}

		err = p.FromXDR(xdrEnv.Operations()[1])
		if assert.NoError(t, err) {
			assert.Equal(t, "", p.SourceAccount, "source accounts should match")
			assert.Equal(t, "GAIH3ULLFQ4DGSECF2AR555KZ4KNDGEKN4AFI4SU2M7B43MGK3QJZNSR", p.Destination, "destination should match")
			assert.Equal(t, "134.0000000", p.Amount, "amount should match")
			assetType, e := p.Asset.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
			assert.Equal(t, "XY", p.Asset.GetCode(), "Asset code should match")
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", p.Asset.GetIssuer(), "Asset issuer should match")
		}
	}
}

func TestPathPaymentFromXDR(t *testing.T) {
	txeB64 := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAAql0AAAADAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAACAAAAAAAAAAAF9eEAAAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAAAAAAAAAmJaAAAAAAQAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAAAAAAAEuFVmYAAAAQF2kLUL/RoFIy1cmt+GXdWn2tDUjJYV3YwF4A82zIBhqYSO6ogOoLPNRt3w+IGCAgfR4Q9lpax+wCXWoQERHSw4="

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var pp PathPayment
		err = pp.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "", pp.SourceAccount, "source accounts should match")
			assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", pp.Destination, "destination should match")
			assert.Equal(t, "1.0000000", pp.DestAmount, "DestAmount should match")
			assert.Equal(t, "10.0000000", pp.SendMax, "SendMax should match")
			assert.Equal(t, true, pp.DestAsset.IsNative(), "DestAsset should be native")
			assert.Equal(t, 1, len(pp.Path), "Number of paths should be 1")
			assetType, e := pp.Path[0].GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
			assert.Equal(t, "ABCD", pp.Path[0].GetCode(), "Asset code should match")
			assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", pp.Path[0].GetIssuer(), "Asset issuer should match")
		}
	}
}

func TestManageSellOfferFromXDR(t *testing.T) {
	txeB64 := "AAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAABkAAADXYAAAABAAAAAAAAAAAAAAACAAAAAQAAAABooIjdquKgVwBESJjRJ3KWgA8pcgUbc3aqRLJLugjdMQAAAAMAAAAAAAAAAkFCQ1hZWgAAAAAAAAAAAABooIjdquKgVwBESJjRJ3KWgA8pcgUbc3aqRLJLugjdMQAAAACy0F4AAAAABQAAAAEAAAAAAAAAAAAAAAAAAAADAAAAAUFCQwAAAAAAaKCI3arioFcAREiY0SdyloAPKXIFG3N2qkSyS7oI3TEAAAAAAAAAAO5rKAAAAAAFAAAAAQAAAAAAAAAAAAAAAAAAAAA="

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var mso ManageSellOffer
		err = mso.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mso.SourceAccount, "source accounts should match")
			assert.Equal(t, int64(0), mso.OfferID, "OfferID should match")
			assert.Equal(t, "300.0000000", mso.Amount, "Amount should match")
			assert.Equal(t, xdr.Price{5, 1}, mso.Price, "Price should match")
			assert.Equal(t, true, mso.Selling.IsNative(), "Selling should be native")
			assetType, e := mso.Buying.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum12, assetType, "Asset type should match")
			assert.Equal(t, "ABCXYZ", mso.Buying.GetCode(), "Asset code should match")
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mso.Buying.GetIssuer(), "Asset issuer should match")
		}

		err = mso.FromXDR(xdrEnv.Operations()[1])
		if assert.NoError(t, err) {
			assert.Equal(t, "", mso.SourceAccount, "source accounts should match")
			assert.Equal(t, int64(0), mso.OfferID, "OfferID should match")
			assert.Equal(t, "400.0000000", mso.Amount, "Amount should match")
			assert.Equal(t, xdr.Price{5, 1}, mso.Price, "Price should match")
			assert.Equal(t, true, mso.Buying.IsNative(), "Buying should be native")
			assetType, e := mso.Selling.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
			assert.Equal(t, "ABC", mso.Selling.GetCode(), "Asset code should match")
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mso.Selling.GetIssuer(), "Asset issuer should match")
		}

	}
}

func TestManageBuyOfferFromXDR(t *testing.T) {
	txeB64 := "AAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAABkAAADXYAAAABAAAAAAAAAAAAAAACAAAAAQAAAABooIjdquKgVwBESJjRJ3KWgA8pcgUbc3aqRLJLugjdMQAAAAwAAAAAAAAAAkFCQ1hZWgAAAAAAAAAAAABooIjdquKgVwBESJjRJ3KWgA8pcgUbc3aqRLJLugjdMQAAAAA7msoAAAAAAQAAAAIAAAAAAAAAAAAAAAAAAAAMAAAAAUFCQwAAAAAAaKCI3arioFcAREiY0SdyloAPKXIFG3N2qkSyS7oI3TEAAAAAAAAAALLQXgAAAAADAAAABQAAAAAAAAAAAAAAAAAAAAA="

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var mbo ManageBuyOffer
		err = mbo.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mbo.SourceAccount, "source accounts should match")
			assert.Equal(t, int64(0), mbo.OfferID, "OfferID should match")
			assert.Equal(t, "100.0000000", mbo.Amount, "Amount should match")
			assert.Equal(t, xdr.Price{1, 2}, mbo.Price, "Price should match")
			assert.Equal(t, true, mbo.Selling.IsNative(), "Selling should be native")
			assetType, e := mbo.Buying.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum12, assetType, "Asset type should match")
			assert.Equal(t, "ABCXYZ", mbo.Buying.GetCode(), "Asset code should match")
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mbo.Buying.GetIssuer(), "Asset issuer should match")
		}

		err = mbo.FromXDR(xdrEnv.Operations()[1])
		if assert.NoError(t, err) {
			assert.Equal(t, "", mbo.SourceAccount, "source accounts should match")
			assert.Equal(t, int64(0), mbo.OfferID, "OfferID should match")
			assert.Equal(t, "300.0000000", mbo.Amount, "Amount should match")
			assert.Equal(t, xdr.Price{3, 5}, mbo.Price, "Price should match")
			assert.Equal(t, true, mbo.Buying.IsNative(), "Buying should be native")
			assetType, e := mbo.Selling.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
			assert.Equal(t, "ABC", mbo.Selling.GetCode(), "Asset code should match")
			assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", mbo.Selling.GetIssuer(), "Asset issuer should match")
		}

	}
}

func TestCreatePassiveSellOfferFromXDR(t *testing.T) {
	txeB64 := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAANAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAEAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAAAX14QAAAAABAAAAAQAAAAAAAAAB0odkfgAAAEAgUD7M1UL7x2m2m26ySzcSHxIneOT7/r+s/HLsgWDj6CmpSi1GZrlvtBH+CNuegCwvW09TRZJhp7bLywkaFCoK"

	xdrEnv, err := unmarshalBase64(txeB64)
	if assert.NoError(t, err) {
		var cpo CreatePassiveSellOffer
		err = cpo.FromXDR(xdrEnv.Operations()[0])
		if assert.NoError(t, err) {
			assert.Equal(t, "", cpo.SourceAccount, "source accounts should match")
			assert.Equal(t, "10.0000000", cpo.Amount, "Amount should match")
			assert.Equal(t, xdr.Price{1, 1}, cpo.Price, "Price should match")
			assert.Equal(t, true, cpo.Selling.IsNative(), "Selling should be native")
			assetType, e := cpo.Buying.GetType()
			assert.NoError(t, e)
			assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
			assert.Equal(t, "ABCD", cpo.Buying.GetCode(), "Asset code should match")
			assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", cpo.Buying.GetIssuer(), "Asset issuer should match")
		}
	}
}

func TestSetOptionsFromXDR(t *testing.T) {

	var opSource xdr.AccountId
	err := opSource.SetAddress("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)
	cFlags := xdr.Uint32(0b1101)
	sFlags := xdr.Uint32(0b1111)
	mw := xdr.Uint32(7)
	lt := xdr.Uint32(2)
	mt := xdr.Uint32(4)
	ht := xdr.Uint32(6)
	hDomain := xdr.String32("hcnet.org")
	var skey xdr.SignerKey
	err = skey.SetAddress("GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3")
	assert.NoError(t, err)
	signer := xdr.Signer{
		Key:    skey,
		Weight: xdr.Uint32(4),
	}

	xdrSetOptions := xdr.SetOptionsOp{
		InflationDest: &opSource,
		ClearFlags:    &cFlags,
		SetFlags:      &sFlags,
		MasterWeight:  &mw,
		LowThreshold:  &lt,
		MedThreshold:  &mt,
		HighThreshold: &ht,
		HomeDomain:    &hDomain,
		Signer:        &signer,
	}

	muxSource := opSource.ToMuxedAccount()
	xdrOp := xdr.Operation{
		SourceAccount: &muxSource,
		Body: xdr.OperationBody{
			Type:         xdr.OperationTypeSetOptions,
			SetOptionsOp: &xdrSetOptions,
		},
	}

	var so SetOptions
	err = so.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", so.SourceAccount, "source accounts should match")
		assert.Equal(t, Threshold(7), *so.MasterWeight, "master weight should match")
		assert.Equal(t, Threshold(2), *so.LowThreshold, "low threshold should match")
		assert.Equal(t, Threshold(4), *so.MediumThreshold, "medium threshold should match")
		assert.Equal(t, Threshold(6), *so.HighThreshold, "high threshold should match")
		assert.Equal(t, "hcnet.org", *so.HomeDomain, "Home domain should match")
		assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", so.Signer.Address, "Signer address should match")
		assert.Equal(t, Threshold(4), so.Signer.Weight, "Signer weight should match")
		assert.Equal(t, int(AuthRequired), int(so.SetFlags[0]), "Set AuthRequired flags should match")
		assert.Equal(t, int(AuthRevocable), int(so.SetFlags[1]), "Set AuthRevocable flags should match")
		assert.Equal(t, int(AuthImmutable), int(so.SetFlags[2]), "Set AuthImmutable flags should match")
		assert.Equal(t, int(AuthClawbackEnabled), int(so.SetFlags[3]), "Set AuthClawbackEnabled flags should match")
		assert.Equal(t, int(AuthRequired), int(so.ClearFlags[0]), "Clear AuthRequired flags should match")
		assert.Equal(t, int(AuthImmutable), int(so.ClearFlags[1]), "Clear AuthImmutable flags should match")
		assert.Equal(t, int(AuthClawbackEnabled), int(so.ClearFlags[2]), "Clear AuthClawbackEnabled flags should match")
	}

}

func TestChangeTrustFromXDR(t *testing.T) {
	asset := CreditAsset{Code: "ABC", Issuer: "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3"}
	xdrAsset, err := asset.ToXDR()
	assert.NoError(t, err)
	xdrLimit, err := amount.Parse("5000")
	assert.NoError(t, err)
	changeTrustOp := xdr.ChangeTrustOp{
		Line:  xdrAsset.ToChangeTrustAsset(),
		Limit: xdrLimit,
	}

	var opSource xdr.MuxedAccount
	err = opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)
	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body: xdr.OperationBody{
			Type:          xdr.OperationTypeChangeTrust,
			ChangeTrustOp: &changeTrustOp,
		},
	}

	var ct ChangeTrust
	err = ct.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", ct.SourceAccount, "source accounts should match")
		assetType, e := ct.Line.GetType()
		assert.NoError(t, e)

		assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
		assert.Equal(t, "ABC", ct.Line.GetCode(), "Asset code should match")
		assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", ct.Line.GetIssuer(), "Asset issuer should match")
		assert.Equal(t, "5000.0000000", ct.Limit, "Trustline  limit should match")
	}
}

func TestAllowTrustFromXDR(t *testing.T) {
	xdrAsset := xdr.Asset{}
	allowTrustAsset, err := xdrAsset.ToAssetCode("ABCXYZ")
	assert.NoError(t, err)

	var opSource xdr.MuxedAccount
	err = opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)

	var trustor xdr.AccountId
	err = trustor.SetAddress("GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3")
	assert.NoError(t, err)

	allowTrustOp := xdr.AllowTrustOp{
		Trustor:   trustor,
		Asset:     allowTrustAsset,
		Authorize: xdr.Uint32(xdr.TrustLineFlagsAuthorizedFlag),
	}

	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body: xdr.OperationBody{
			Type:         xdr.OperationTypeAllowTrust,
			AllowTrustOp: &allowTrustOp,
		},
	}

	var at AllowTrust
	err = at.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", at.SourceAccount, "source accounts should match")

		assetType, e := at.Type.GetType()
		assert.NoError(t, e)
		assert.Equal(t, AssetTypeCreditAlphanum12, assetType, "Asset type should match")
		assert.Equal(t, "ABCXYZ", at.Type.GetCode(), "Asset code should match")
		assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", at.Trustor, "Trustor should match")
		assert.Equal(t, true, at.Authorize, "Authorize value should match")
	}
}

func TestAccountMergeFromXDR(t *testing.T) {
	var opSource xdr.MuxedAccount
	err := opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)

	var destination xdr.MuxedAccount
	err = destination.SetEd25519Address("GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3")
	assert.NoError(t, err)

	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body: xdr.OperationBody{
			Type:        xdr.OperationTypeAccountMerge,
			Destination: &destination,
		},
	}

	var am AccountMerge
	err = am.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", am.SourceAccount, "source accounts should match")
		assert.Equal(t, "GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3", am.Destination, "destination accounts should match")
	}
}

func TestInflationFromXDR(t *testing.T) {
	var opSource xdr.MuxedAccount
	err := opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)

	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body:          xdr.OperationBody{Type: xdr.OperationTypeInflation},
	}

	var inf Inflation
	err = inf.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", inf.SourceAccount, "source accounts should match")
	}
}

func TestManageDataFromXDR(t *testing.T) {
	var opSource xdr.MuxedAccount
	err := opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)

	dv := []byte("value")
	xdrdv := xdr.DataValue(dv)
	manageDataOp := xdr.ManageDataOp{
		DataName:  xdr.String64("data"),
		DataValue: &xdrdv,
	}

	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body: xdr.OperationBody{
			Type:         xdr.OperationTypeManageData,
			ManageDataOp: &manageDataOp,
		},
	}

	var md ManageData
	err = md.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", md.SourceAccount, "source accounts should match")
		assert.Equal(t, "data", md.Name, "Name should match")
		assert.Equal(t, "value", string(md.Value), "Value should match")
	}
}

func TestBumpSequenceFromXDR(t *testing.T) {
	var opSource xdr.MuxedAccount
	err := opSource.SetEd25519Address("GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H")
	assert.NoError(t, err)

	bsOp := xdr.BumpSequenceOp{
		BumpTo: xdr.SequenceNumber(45),
	}

	xdrOp := xdr.Operation{
		SourceAccount: &opSource,
		Body: xdr.OperationBody{
			Type:           xdr.OperationTypeBumpSequence,
			BumpSequenceOp: &bsOp,
		},
	}

	var bs BumpSequence
	err = bs.FromXDR(xdrOp)
	if assert.NoError(t, err) {
		assert.Equal(t, "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H", bs.SourceAccount, "source accounts should match")
		assert.Equal(t, int64(45), bs.BumpTo, "BumpTo should match")
	}
}

func testOperationsMarshalingRoundtrip(t *testing.T, operations []Operation, withMuxedAccounts bool) {
	kp1 := newKeypair1()
	accountID := xdr.MustAddress(kp1.Address())
	mx := xdr.MuxedAccount{
		Type: xdr.CryptoKeyTypeKeyTypeMuxedEd25519,
		Med25519: &xdr.MuxedAccountMed25519{
			Id:      0xcafebabe,
			Ed25519: *accountID.Ed25519,
		},
	}
	var sourceAccount SimpleAccount
	if withMuxedAccounts {
		sourceAccount = NewSimpleAccount(mx.Address(), int64(9605939170639898))
	} else {
		sourceAccount = NewSimpleAccount(kp1.Address(), int64(9606132444168199))
	}

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount: &sourceAccount,
			Operations:    operations,
			Preconditions: Preconditions{TimeBounds: NewInfiniteTimeout()},
			BaseFee:       MinBaseFee,
		},
	)
	assert.NoError(t, err)

	var b64 string
	b64, err = tx.Base64()
	assert.NoError(t, err)

	var parsedTx *GenericTransaction
	if withMuxedAccounts {
		parsedTx, err = TransactionFromXDR(b64)
	} else {
		parsedTx, err = TransactionFromXDR(b64)
	}
	assert.NoError(t, err)
	var ok bool
	tx, ok = parsedTx.Transaction()
	assert.True(t, ok)

	for i := 0; i < len(operations); i++ {
		assert.Equal(t, operations[i], tx.Operations()[i])
	}
}

func TestOperationCoverage(t *testing.T) {
	gen := randxdr.NewGenerator()
	for i := 0; i < 10000; i++ {
		op := xdr.Operation{}
		shape := &gxdr.Operation{}
		gen.Next(
			shape,
			[]randxdr.Preset{
				{randxdr.IsDeepAuthorizedInvocationTree, randxdr.SetVecLen(0)},
				{
					randxdr.FieldEquals("body.revokeSponsorshipOp.ledgerKey.type"),
					randxdr.SetU32(
						gxdr.ACCOUNT.GetU32(),
						gxdr.TRUSTLINE.GetU32(),
						gxdr.OFFER.GetU32(),
						gxdr.DATA.GetU32(),
						gxdr.CLAIMABLE_BALANCE.GetU32(),
					),
				},
			},
		)
		assert.NoError(t, gxdr.Convert(shape, &op))

		_, err := operationFromXDR(op)
		if !assert.NoError(t, err) {
			t.Log(goxdr.XdrToString(shape))
		}
	}
}