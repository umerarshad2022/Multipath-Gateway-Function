# Multipath-Gateway-Function
Multipath Gateway Function for a Hybrid 5G and WiFi Network
# Bi-Directional Communication with NAT and MPTCP

![Architecture Overview](./GNS-setup.png)


## Overview
This project demonstrates a bi-directional communication architecture involving multiple hosts, a **Multipath Function (MPF) Gateway**, and an external server. The architecture uses **Network Address Translation (NAT)** and **Multipath TCP (MPTCP)** for traffic routing between internal and external networks. The packets sent from **Host A** are captured on the MPF Gateway using the **PCAP library** and forwarded to the external server, which communicates with **Host B**.

### Key Features:
- **Bi-directional communication** between Host A and Host B through the external server.
- **NAT** implemented at the MPF Gateway node for address translation between internal (Host A) and external networks.
- **MPTCP** (Multipath TCP) enabled between the MPF Gateway and the external server to enhance communication over multiple network interfaces.
- **PCAP library** used to capture traffic from Host A at the MPF Gateway.

## Architecture

### 1. Host A (Internal Network)
- **Role**: Host A is part of the internal network and communicates with the external server through the MPF Gateway.
- **Communication**: Packets sent from Host A are captured on the MPF Gateway using the **PCAP library**. These packets are then translated using **NAT** and forwarded to the external server using MPTCP. Host A can also receive responses from the external server via the same NAT process.
- **Bi-Directional Traffic**: Host A can send and receive traffic to/from Host B via the external server. The MPF Gateway handles NAT and MPTCP communication between Host A and the external server.

### 2. MPF Gateway (Gateway Node)
- **Role**: The **MPF Gateway** is a central node that:
  - Captures packets from **Host A** using the **PCAP library**.
  - Implements **NAT** (Network Address Translation): The gateway translates Host A’s internal IP address to the gateway’s external IP before forwarding the traffic to the external server.
  - Uses **MPTCP (Multipath TCP)**: The gateway communicates with the external server over multiple network interfaces (e.g., `ens7`, `ens8`), using MPTCP to enhance reliability and performance by sending traffic across multiple paths.
- **Bi-Directional Traffic Handling**:
  - **Outbound**: The MPF Gateway captures traffic from Host A using **PCAP**, applies **NAT**, and forwards the translated packets to the external server using MPTCP.
  - **Inbound**: The MPF Gateway receives traffic from Host B (through the external server), performs reverse **NAT**, and forwards the traffic to Host A.

### 3. External Server
- **Role**: The external server acts as an intermediary between Host A and Host B. It receives traffic from Host A via the MPF Gateway and forwards it to Host B.
- **Communication**: The external server communicates with the MPF Gateway using MPTCP and forwards traffic between Host A and Host B. Any responses from Host B are sent back to the MPF Gateway.
- **MPTCP**: The external server uses MPTCP to handle multiple TCP subflows coming from the MPF Gateway, providing enhanced communication across multiple network interfaces.

### 4. Host B (External Network)
- **Role**: Host B is part of the external network and communicates with the external server. It receives traffic from Host A (via the external server and the MPF Gateway) and can send traffic back through the same path.
- **Communication**: Host B is directly connected to the external server. It receives traffic from the external server and responds back through the server, which forwards the traffic to the MPF Gateway and then to Host A.
- **Bi-Directional Traffic**: Host B can send and receive traffic to/from Host A through the external server.

## Communication Flow
### Outbound (Host A to Host B):
1. **Host A** sends traffic destined for **Host B**.
2. **MPF Gateway** captures the packet using the **PCAP library**, performs **NAT**, and translates the internal IP address of Host A to the gateway’s external IP.
3. **MPF Gateway** forwards the traffic to the **External Server** using **MPTCP** over multiple network interfaces.
4. **External Server** forwards the traffic to **Host B**, which is directly connected to the server.
5. **Host B** receives the traffic from the external server.

### Inbound (Host B to Host A):
1. **Host B** sends traffic to **Host A** via the **External Server**.
2. **External Server** forwards the traffic to the **MPF Gateway** using **MPTCP**.
3. **MPF Gateway** performs reverse **NAT**, translating the external IP back to Host A’s internal IP.
4. **MPF Gateway** forwards the traffic to **Host A**.

## Technologies Used
- **Go**: The primary language used to implement NAT and MPTCP socket management.
- **pcap/gopacket**: Used for **capturing packets** from Host A at the MPF Gateway.
- **Multipath TCP (MPTCP)**: Allows the MPF Gateway to communicate with the external server over multiple network paths for improved reliability and bandwidth aggregation.
- **Network Address Translation (NAT)**: Implemented at the MPF Gateway to allow communication between the internal host (Host A) and the external server, translating IP addresses for correct routing.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


