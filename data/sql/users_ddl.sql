CREATE TABLE IF NOT EXISTS user_dim (
      user_id               INTEGER         PRIMARY KEY
    , username              VARCHAR(64)     NOT NULL
    , password              VARCHAR(64)     NOT NULL
)
;
