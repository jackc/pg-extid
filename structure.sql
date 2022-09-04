create extension if not exists pgcrypto;

create table extid_types (
	prefix text not null primary key,
	secret_key bytea not null default gen_random_bytes(16)
);

create function encode_extid(prefix text, id bigint)
returns text
stable
language sql as $$
	select extid_types.prefix || '_' || encode(encrypt(int8send(id), secret_key, 'aes-ecb/pad:none'), 'hex')
	from extid_types
	where extid_types.prefix = $1;
$$;

create function decode_extid(extid text)
returns bigint
stable
language sql as $$
	select ('x' || left(encode(decrypt(decode(split_part($1, '_', 2), 'hex'), secret_key, 'aes-ecb/pad:none'), 'hex'), 16))::bit(64)::bigint
	from extid_types
	where extid_types.prefix = split_part($1, '_', 1);
$$;
