-- 000001_init.up.sql
-- StudyPilot P0 schema（SQLite）
-- 约定：日期为本地字符串 YYYY-MM-DD；时间为当日分钟数；周以 week_start（周一）为键。

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE goals (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT    NOT NULL,
    goal_type    TEXT    NOT NULL CHECK (goal_type IN ('day','week')),
    goal_mode    TEXT    NOT NULL DEFAULT 'daily' CHECK (goal_mode IN ('daily','once')),
    date         TEXT,
    week_start   TEXT,
    checkin_days INTEGER NOT NULL DEFAULT 5 CHECK (checkin_days BETWEEN 1 AND 7),
    status       TEXT    NOT NULL DEFAULT 'active' CHECK (status IN ('active','done','archived')),
    note         TEXT    NOT NULL DEFAULT '',
    sort         INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    CHECK (
        (goal_type = 'day'  AND date IS NOT NULL AND week_start IS NULL) OR
        (goal_type = 'week' AND week_start IS NOT NULL AND date IS NULL)
    ),
    CHECK (
        goal_mode = 'daily' OR (goal_mode = 'once' AND goal_type = 'week')
    )
);

CREATE TABLE goal_checkins (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    goal_id      INTEGER NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    checkin_date TEXT    NOT NULL,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    UNIQUE (goal_id, checkin_date)
);

CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT    NOT NULL,
    note         TEXT    NOT NULL DEFAULT '',
    parent_id    INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    goal_id      INTEGER REFERENCES goals(id) ON DELETE SET NULL,
    status       TEXT    NOT NULL DEFAULT 'todo' CHECK (status IN ('todo','doing','done')),
    priority     INTEGER NOT NULL DEFAULT 1 CHECK (priority BETWEEN 0 AND 3),
    due_date     TEXT,
    est_minutes  INTEGER CHECK (est_minutes IS NULL OR est_minutes > 0),
    tags         TEXT    NOT NULL DEFAULT '[]',
    sort         INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE task_events (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id       INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type    TEXT    NOT NULL CHECK (event_type IN ('created','done','reopen')),
    occurred_date TEXT    NOT NULL,
    occurred_at   TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE blocks (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    block_date TEXT    NOT NULL,
    start_min  INTEGER NOT NULL CHECK (start_min BETWEEN 0 AND 1439),
    end_min    INTEGER NOT NULL CHECK (end_min BETWEEN 1 AND 1440),
    title      TEXT    NOT NULL,
    task_id    INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    color      TEXT    NOT NULL DEFAULT '',
    note       TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    CHECK (end_min > start_min)
);

CREATE TABLE reviews (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    week_start     TEXT    NOT NULL UNIQUE,
    title          TEXT    NOT NULL DEFAULT '',
    content_md     TEXT    NOT NULL DEFAULT '',
    stats_snapshot TEXT    NOT NULL DEFAULT '{}',
    created_at     TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE INDEX idx_goals_date       ON goals(date);
CREATE INDEX idx_goals_week       ON goals(week_start);
CREATE INDEX idx_checkins_date    ON goal_checkins(checkin_date);
CREATE INDEX idx_tasks_parent     ON tasks(parent_id);
CREATE INDEX idx_tasks_goal       ON tasks(goal_id);
CREATE INDEX idx_tasks_status_due ON tasks(status, due_date);
CREATE INDEX idx_events_date      ON task_events(occurred_date);
CREATE INDEX idx_events_task      ON task_events(task_id, occurred_date);
CREATE INDEX idx_blocks_date      ON blocks(block_date);
CREATE INDEX idx_blocks_task      ON blocks(task_id);