-- CREATE TABLE IF NOT EXISTS auth_user_policy (
--       user_policy_id        INTEGER         PRIMARY KEY
--     , subject               VARCHAR(64)     NOT NULL
--     , object                VARCHAR(64)     NOT NULL
--     , action                VARCHAR(64)
--     , effect                VARCHAR(64)     DEFAULT 'allow'
-- )
-- ;

-- CREATE TABLE IF NOT EXISTS auth_role_policy (
--       role_policy_id        INTEGER         PRIMARY KEY
--     , subject               VARCHAR(64)     NOT NULL
--     , object                VARCHAR(64)     NOT NULL
--     , action                VARCHAR(64)
--     , effect                VARCHAR(64)     DEFAULT 'allow'
-- )
-- ;

-- CREATE TABLE IF NOT EXISTS auth_role_dim (
--       role_dim_id           INTEGER         PRIMARY KEY
--     , role_name             INTEGER         UNIQUE NOT NULL
-- )
-- ;

-- CREATE TABLE IF NOT EXISTS auth_user_role_map_policy (
--       map_policy_id         INTEGER         PRIMARY KEY
--     , subject               VARCHAR(64)     NOT NULL
--     , object                VARCHAR(64)     NOT NULL
-- )
-- ;

CREATE VIEW casbin_rules AS
    SELECT
          aup.user_policy_id            AS policy_id
        , 'u' || aup.subject            AS subject
        , aup.object                    AS object
        , aup.action                    AS action
        , aup.effect                    AS effect
    FROM
        auth_user_policy    AS aup
    UNION
    SELECT
          arp.role_policy_id            AS policy_id
        , 'r' || arp.subject            AS subject
        , arp.object                    AS object
        , arp.action                    AS action
        , arp.effect                    AS effect
    FROM
        auth_role_policy    AS arp
    UNION
    SELECT
          aurmp.map_policy_id           AS policy_id
        , 'u' || aurmp.subject          AS subject
        , 'r' || aurmp.object           AS object
        , NULL                          AS action
        , NULL                          AS effect
    FROM
        auth_user_role_map_policy   AS aurmp
;
