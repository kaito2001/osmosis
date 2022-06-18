package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v7/app/params"
)

func TestMsgSwapExactAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		properMsg := MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []SwapAmountInRoute{{
				PoolId:        0,
				TokenOutDenom: "test",
			}, {
				PoolId:        1,
				TokenOutDenom: "test2",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	createMultihopMsg := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		properMsg := MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []SwapAmountInRoute{{
				PoolId:        0,
				TokenOutDenom: "test2",
			}, {
				PoolId:        1,
				TokenOutDenom: "test1",
			}, {
				PoolId:        2,
				TokenOutDenom: "test",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgMultihop := createMultihopMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		// do nothing
		return msg
	})

	require.Equal(t, msgMultihop.Route(), RouterKey)
	require.Equal(t, msgMultihop.Type(), "swap_exact_amount_in")
	signersMultOut := msgMultihop.GetSigners()
	require.Equal(t, len(signersMultOut), 1)
	require.Equal(t, signersMultOut[0].String(), addr1)

	tests := []struct {
		name             string
		msg              SybilResistantFee
		expectDenomPath  []string
		expectPoolIdPath []uint64
		expectToken      sdk.Coin
		expectPass       bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = []SwapAmountInRoute{}
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectDenomPath:  []string{"test", "1"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.Coin{},
			expectPass:       false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.ZeroInt()),
			expectPass:       false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.Coin{},
			expectPass:       false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "proper swap in multihop",
			msg: createMultihopMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test1", "test2"},
			expectPoolIdPath: []uint64{0, 1, 2},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       true,
		},
	}

	for _, test := range tests {
		msg, ok := test.msg.(MsgSwapExactAmountIn)
		require.True(t, ok)

		if test.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", test.name)
		}

		denomPath := test.msg.GetTokenDenomsOnPath()
		poolIdPath := test.msg.GetPoolIdOnPath()
		token := test.msg.GetTokenToFee()

		require.Equalf(t, test.expectToken, token, "test: %v", test.name)
		require.ElementsMatchf(t, test.expectDenomPath, denomPath, "test: %v", test.name)
		require.ElementsMatchf(t, test.expectPoolIdPath, poolIdPath, "test: %v", test.name)
	}
}

func TestMsgSwapExactAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		properMsg := MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []SwapAmountOutRoute{{
				PoolId:       0,
				TokenInDenom: "test",
			}, {
				PoolId:       1,
				TokenInDenom: "test2",
			}},
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			TokenInMaxAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	createMultihopMsg := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		properMsg := MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []SwapAmountOutRoute{{
				PoolId:       2,
				TokenInDenom: "test2",
			}, {
				PoolId:       1,
				TokenInDenom: "test1",
			}, {
				PoolId:       0,
				TokenInDenom: "test",
			}},
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			TokenInMaxAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgMultihop := createMultihopMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		// do nothing
		return msg
	})

	require.Equal(t, msgMultihop.Route(), RouterKey)
	require.Equal(t, msgMultihop.Type(), "swap_exact_amount_out")
	signersMultOut := msgMultihop.GetSigners()
	require.Equal(t, len(signersMultOut), 1)
	require.Equal(t, signersMultOut[0].String(), addr1)

	tests := []struct {
		name             string
		msg              SybilResistantFee
		expectDenomPath  []string
		expectPoolIdPath []uint64
		expectToken      sdk.Coin
		expectPass       bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = []SwapAmountOutRoute{}
				return msg
			}),

			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),

			expectDenomPath:  []string{"1", "test"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.Coin{},
			expectPass:       false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.ZeroInt()),
			expectPass:       false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.Coin{},
			expectPass:       false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),

			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),

			expectDenomPath:  []string{"test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       false,
		},
		{
			name: "proper swap out multihop",
			msg: createMultihopMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test2", "test1"},
			expectPoolIdPath: []uint64{2, 1, 0},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass:       true,
		},
	}

	for _, test := range tests {
		msg, ok := test.msg.(MsgSwapExactAmountOut)
		require.True(t, ok)

		if test.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", test.name)
		}

		denomPath := test.msg.GetTokenDenomsOnPath()
		poolIdPath := test.msg.GetPoolIdOnPath()
		token := test.msg.GetTokenToFee()

		require.Equalf(t, test.expectToken, token, "test: %v", test.name)
		require.ElementsMatchf(t, test.expectDenomPath, denomPath, "test: %v", test.name)
		require.ElementsMatchf(t, test.expectPoolIdPath, poolIdPath, "test: %v", test.name)
	}
}

func TestMsgJoinPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinPool) MsgJoinPool) MsgJoinPool {
		properMsg := MsgJoinPool{
			Sender:         addr1,
			PoolId:         1,
			ShareOutAmount: sdk.NewInt(10),
			TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinPool) MsgJoinPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token max in' can pass",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token max in' can pass 2",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = sdk.Coins{}
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgExitPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitPool) MsgExitPool) MsgExitPool {
		properMsg := MsgExitPool{
			Sender:        addr1,
			PoolId:        1,
			ShareInAmount: sdk.NewInt(10),
			TokenOutMins:  sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitPool) MsgExitPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token min out' can pass",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token min out' can pass 2",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = sdk.Coins{}
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgJoinSwapExternAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
		properMsg := MsgJoinSwapExternAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			ShareOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_swap_extern_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinSwapExternAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgJoinSwapShareAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
		properMsg := MsgJoinSwapShareAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenInDenom:     "test",
			ShareOutAmount:   sdk.NewInt(100),
			TokenInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_swap_share_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinSwapShareAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgExitSwapExternAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
		properMsg := MsgExitSwapExternAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			ShareInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_extern_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitSwapExternAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgExitSwapShareAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
		properMsg := MsgExitSwapShareAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenOutDenom:     "test",
			ShareInAmount:     sdk.NewInt(100),
			TokenOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_share_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitSwapShareAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}
