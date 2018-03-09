var RLP = artifacts.require("./libraries/RLP.sol");

module.exports = function(deployer) {
  deployer.deploy(RLP);
};
