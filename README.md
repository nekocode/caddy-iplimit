Usage:

```
iplimit max_amount max_age
```

* `max_amount` is the maximum number of IPs
* `max_age` is the maximum inactive time per IP

Example:

```
:2016 {
  iplimit 1 10s
}
```
