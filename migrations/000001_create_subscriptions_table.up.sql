CREATE TABLE IF NOT EXISTS subscriptions
(
    id varchar not null,
    name varchar unique not null,
    started date not null,
    expire date not null,
    price int not null,
    PRIMARY KEY (id, name)
);
ALTER TABLE subscriptions
    ADD CONSTRAINT unique_user_subscription UNIQUE (id, name);

INSERT INTO subscriptions (id, name, started, expire, price) VALUES
    ('testID1', 'youtube', '2025-08-11', '2025-12-08', 300),
    ('testID1', 'twitch', '2025-09-25', '2025-12-29', 500),
    ('testID2', 'yandexMusic', '2024-01-01', '2025-12-13', 400)
;