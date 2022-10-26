# pg-extid

It can be valuable to internally use a serial integer as an ID without revealing that ID to the outside world. pg-extid
uses AES-128 to convert to and from an external ID that cannot feasibly be decoded without the secret key.

This prevents outsiders from quantifying the usage of your application by observing the rate of increase of IDs as well
as provides protection against brute force crawling of all resources.

## Installation

Run `structure.sql` in your database.

## Example Usage

```sql
insert into extid_types (prefix) values ('user');
-- => INSERT 0 1

select encode_extid('user', 123);
-- => user_d1001653c9461a3e97106d386d0bb4df

select decode_extid('user_d1001653c9461a3e97106d386d0bb4df');
-- => 123
```

## Performance

Encoding an ID takes ~10µs and decoding an external ID takes ~15µs on a 2019 MacBook Pro.

```sql
select encode_extid('user', n) from generate_series(1, 100000) n;
-- Time: 1070.191 ms (00:01.070)

select decode_extid(encode_extid('user', n)) from generate_series(1, 100000) n;
-- Time: 2508.360 ms (00:02.508)
```

## Other Implementations

* [Go](https://github.com/jackc/go-extid)
* [Ruby](https://github.com/jackc/ruby-extid)
