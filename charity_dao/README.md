# Sample Hardhat Project

This project demonstrates a basic Hardhat use case. It comes with a sample contract, a test for that contract, and a Hardhat Ignition module that deploys that contract.

Try running some of the following tasks:

```shell
npx hardhat help
npx hardhat test
REPORT_GAS=true npx hardhat test
npx hardhat node
npx hardhat ignition deploy ./ignition/modules/Lock.js
```
'''deploy to docker
git pull
docker-compose build
docker-compose up -d
'''

```
npx hardhat console --network localhost

// Lấy danh sách các signer (ví giả lập)
const [owner, addr1] = await ethers.getSigners();
// Kết nối tới contract đã deploy (copy đúng địa chỉ từ log của bạn)
const vault = await ethers.getContractAt("CharityVault", "0x5FbDB23...");
// Gửi 1 ETH từ addr1 tới contract (donate)
await vault.connect(addr1).donate({ value: ethers.parseEther("1.0") });
// Kiểm tra số dư contract
(await ethers.provider.getBalance(vault.target || vault.address)).toString()
// Admin (owner) rút 0.5 ETH
await vault.connect(owner).withdraw(ethers.parseEther("0.5"));
// Kiểm tra lại số dư contract sau khi rút
(await ethers.provider.getBalance(vault.target || vault.address)).toString()
// Xem số dư addr1 (người donate)
(await ethers.provider.getBalance(addr1.address)).toString()

```