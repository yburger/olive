INSERT INTO shows (show_id, enable, platform, room_id, streamer_name, out_tmpl, parser, save_dir, post_cmds, split_rule, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'true', 'bilibili', '21852', 'old-tomato', '[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv', 'flv', '', '', '', '2022-10-05 00:00:00', '2022-10-05 00:00:00')
	ON CONFLICT DO NOTHING;
