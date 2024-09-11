MONIKER=$1
# if empty, default to "testnet"
if [ -z "$MONIKER" ]; then
  MONIKER="mynode"
fi
export CHAIN_ID=testnet
export VALKEY=val1
export STAKE_HOME=./build/simapp_home
export KEYRING=test
ALICE_MNEMONIC="history perfect across group seek acoustic delay captain sauce audit carpet tattoo exhaust green there giant cluster want pond bulk close screen scissors remind"
BOB_MNEMONIC="limb sister humor wisdom elephant weasel beyond must any desert glance stem reform soccer include chest chef clerk call popular display nerve priority venture"
VAL1_MNEMONIC="note base stone list envelope tail start forget alarm acoustic cook occur divert giant bike curtain chase shuffle fade glow capital slot file provide"
VAL2_MNEMONIC="strike tower consider despair bridge diesel clay celery violin base hello ride they weather tunnel elite truth oblige spot hen wise flag pet battle"
VAL3_MNEMONIC="canvas human require month loan oak december blame grit palm slice error absorb total spice autumn trouble soda repeat shove quit bid forward organ"
VAL4_MNEMONIC="grant raw marine drink text dove flat waste wish buzz output hand merge cluster civil clog stay alert silent reunion idea cake village almost"
SUPER_ADMIN_MNEMONIC="expect peace defense conduct virtual flight flip unit equip solve broccoli protect shed group else useless tree such tornado minimum decade tower warfare galaxy"

rm -Rf ${STAKE_HOME}

simd init ${MONIKER} --chain-id=${CHAIN_ID} --home ${STAKE_HOME}

# mint to validator
echo $SUPER_ADMIN_MNEMONIC | simd keys add super-admin --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $ALICE_MNEMONIC | simd keys add alice --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $BOB_MNEMONIC | simd keys add bob --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $VAL1_MNEMONIC | simd keys add val1 --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $VAL2_MNEMONIC | simd keys add val2 --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $VAL3_MNEMONIC | simd keys add val3 --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}
echo $VAL4_MNEMONIC | simd keys add val4 --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING}

simd add-genesis-account $(simd keys show -a val1 --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 11000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a val2 --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 11000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a val3 --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 11000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a val4 --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 11000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a alice --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a bob --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a super-admin --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000000stake --keyring-backend ${KEYRING} --home ${STAKE_HOME}

simd gentx ${VALKEY} 1000000000000stake --chain-id=${CHAIN_ID} --keyring-backend=test --home ${STAKE_HOME}
simd collect-gentxs --home ${STAKE_HOME}