-- Version: 0.5
-- Description: Create table shows
CREATE TABLE shows (
	show_id       UUID,
	enable        BOOLEAN,
	platform   	  TEXT,
	room_id       TEXT,
	streamer_name TEXT,
	out_tmpl      TEXT,
	parser        TEXT,
	save_dir      TEXT,
	post_cmds     TEXT,
	split_rule    TEXT,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,
	
	PRIMARY KEY (show_id)
);

CREATE TABLE configs (
	key  TEXT,
	value  TEXT,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,
	
	PRIMARY KEY (key)
);
