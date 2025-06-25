export STAKE_HOME=/opt/build/simapp_home/
simd start --home ${STAKE_HOME} --minimum-gas-prices=1.25stake --rpc.laddr "tcp://0.0.0.0:26657"