# Trusted OceanBase RPM signing keys

OBShell embeds these public keys and verifies the `obshell` RPM in-process when
OBShell upgrades itself. OceanBase, OBProxy, SeekDB, and other RPM flows retain
their existing behavior. Verification does not read the host RPM database,
invoke the `rpm` command, or download keys at runtime.

The files were downloaded from the official OceanBase RPM repository on
2026-08-12. Their SHA-256 digests and primary key fingerprints are:

| File | SHA-256 | Primary key fingerprint |
| --- | --- | --- |
| `RPM-GPG-KEY-OceanBase` | `c0a9f872be9a99a5f28ea9fa3332df0bb345e9abfcec53b0ff054fa57f6bb8d3` | `A109 E086 B8F1 D712 A255 5CD4 3712 65DA 09E1 CEBF` |
| `RPM-GPG-KEY-OceanBase-el7` | `a8f7ec5059e8ef476d41cc7a1cb4f40ec1d35afa50fff67250104b6d576b2586` | `A74F E069 A852 A639 A241 2347 283B CA09 6347 6B06` |
| `RPM-GPG-KEY-OceanBase-old` | `dff33e8dcee747a23c285eaaea1d7dea14175000b726e9452bd90c431c21aafe` | `EF7D E8E3 6987 B60C ACF9 9A53 2FF8 45A6 E9B4 A7AA` |

Sources:

- <https://mirrors.oceanbase.com/RPM-GPG-KEY-OceanBase>
- <https://mirrors.oceanbase.com/RPM-GPG-KEY-OceanBase-el7>
- <https://mirrors.oceanbase.com/RPM-GPG-KEY-OceanBase-old>

Key rotation must be performed as a reviewed source change. Add the new public
key before release artifacts begin using its private counterpart, verify a real
release RPM in tests, and remove retired keys according to the release support
policy.
