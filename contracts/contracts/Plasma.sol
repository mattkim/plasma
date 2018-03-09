pragma solidity ^0.4.17;

import './libraries/SafeMath.sol';
import './libraries/RLP.sol';


contract Plasma {
    using SafeMath for uint256;
    using RLP for bytes;
    using RLP for RLP.RLPItem;
    using RLP for RLP.Iterator;

    event Deposit(address sender, uint value);

    address public authority;
    mapping(uint256 => childBlock) public childChain;
    uint256 public currentChildBlock;

    struct childBlock {
        bytes32 root;
        uint256 created_at;
    }

    function Plasma() {
        authority = msg.sender;
        currentChildBlock = 1;
    }

    function deposit() public payable {
        Deposit(msg.sender, msg.value);
    }

    function deposit2(bytes txBytes) public payable {
        bytes32 root = createSimpleMerkleRoot(txBytes);

        childChain[currentChildBlock] = childBlock({
            root: root,
            created_at: block.timestamp
        });

        currentChildBlock = currentChildBlock.add(1);

        // Implies that the sender is always the recipient of the deposit.
        Deposit(msg.sender, msg.value);
    }

    function createSimpleMerkleRoot(bytes txBytes) returns (bytes32) {
        bytes32 zeroBytes;
        // TODO: Why is this 130 added to the end again?
        // This is the left and right leaf of this hash.
        bytes32 root = keccak256(keccak256(txBytes), new bytes(130));
        for (uint i = 0; i < 16; i++) {
            root = keccak256(root, zeroBytes);
            zeroBytes = keccak256(zeroBytes, zeroBytes);
        }

        return root;
    }
}
