var RLP = artifacts.require("./libraries/RLP.sol");
var SafeMath = artifacts.require("./libraries/SafeMath.sol");

module.exports = function(deployer) {
  // TODO: I might need to link this to the Plasma contract.
  deployer.deploy(RLP);
  deployer.deploy(SafeMath);
};
