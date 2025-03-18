create table messages
(
    message_id varchar(512) not null,
    header varchar(128),
    body varchar(512),
    created_at date,
    updated_at date
);