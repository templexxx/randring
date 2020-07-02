# randring

Package randring provides a ring buckets for multi-producer & one-consumer
which will drop messages if buckets full and won't guarantee order.

randring only cares about memory corruption.
