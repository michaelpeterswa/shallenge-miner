CREATE TABLE IF NOT EXISTS
    shallengeminer.results (
        name text NOT NULL,
        nonce text NOT NULL,
        sha256 text NOT NULL,
        quality float NOT NULL,
        created_at timestamp default current_timestamp
    );