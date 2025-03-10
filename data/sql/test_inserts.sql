-- TEST DATA

-- INSERT INTO user_dim (user_id, username, password)
-- VALUES
--     (1, 'Ray',   'bestpass'),
--     (2, 'Tadej', 'goodpass'),
--     (3, 'Petar', 'nopass')
-- ;

-- hint: some usernames must be forbidden because we will use them for role names
INSERT INTO auth_role_dim (role_dim_id, role_name)
VALUES
    (1, 'B_admin'),
    (2, 'B_minion')
;

INSERT INTO auth_role_policy (role_policy_id, subject, object, action, effect)
VALUES
    -- define base role policies
    -- additionaly there can be a table for "objects" - give every object it's own ID
        -- question: how to handle it to still be readable? It might end up looking like auth_user_role_map_policy inserts
            -- perhaps we can store "numbers" in policy tables, but at every point where we are looking at the data we need to join with dim table to fetch the names for subjects and objects
            -- e.g. Admin panel will show role names, user names and object/resource names, but before it gets inserted into the DB we will have a transform function to map names to IDs
        -- we could potentially also transform "allow" = 1 and "deny" = 0 ?

    -- B_admin
    (100, 1, 'report_text',                 'read',     'allow'),
    (101, 1, 'inputbox_client_name',        'read',     'allow'),
    (102, 1, 'inputbox_client_name',        'write',    'deny'),
    (103, 1, 'inputbox_time_spent',         'read',     'allow'),
    (104, 1, 'inputbox_time_spent',         'write',    'deny'),
    (105, 1, 'admin_text',                  'read',     'allow'),
    -- B_minion
    (200, 2, 'report_text',                 'read',     'allow'),
    (201, 2, 'inputbox_client_name',        'read',     'allow'),
    (202, 2, 'inputbox_client_name',        'write',    'allow'),
    (203, 2, 'inputbox_time_spent',         'read',     'allow'),
    (204, 2, 'inputbox_time_spent',         'write',    'allow'),
    (205, 2, 'admin_text',                  'read',     'deny')
;

INSERT INTO auth_user_policy (user_policy_id, subject, object, action, effect)
VALUES
    -- deny report text from Petar
    (1, 3, 'report_text',                   'read',     'deny')
;

INSERT INTO auth_user_role_map_policy (map_policy_id, subject, object)
VALUES
    -- very much unreadable without mapping with dim tables. Hackproof?
    (1, 1, 1),
    (2, 2, 2),
    (3, 3, 2)
;

-- Kristine & Preston users