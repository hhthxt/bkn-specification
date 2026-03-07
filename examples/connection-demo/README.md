# Connection Demo

Demonstrates the optional **connection** capability: multiple objects sharing one data-source connection.

## Structure

| File | Purpose |
|------|---------|
| `connections/erp_db.bkn` | Defines ERP database connection (postgres) |
| `objects/material.bkn` | Material object, references `connection \| erp_db` |
| `objects/inventory.bkn` | Inventory object, references `connection \| erp_db` (shares same connection) |
| `objects/legacy_view.bkn` | Uses traditional `data_view` (compatibility) |

## Usage

```python
from bkn import load_network

net = load_network("examples/connection-demo/index.bkn")
conn = net.get_connection("erp_db")
# conn.config.conn_type, conn.config.endpoint, conn.config.secret_ref
```

```go
net, _ := bkn.LoadNetwork("examples/connection-demo/index.bkn")
conn := net.GetConnection("erp_db")
// conn.Config.ConnType, conn.Config.Endpoint, conn.Config.SecretRef
```

## Compatibility

- Networks without `connection` continue to work unchanged
- `connection` is optional; no migration required
