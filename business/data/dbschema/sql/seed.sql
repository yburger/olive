-- INSERT INTO shows (show_id, enable, platform, room_id, streamer_name, out_tmpl, parser, save_dir, post_cmds, split_rule, date_created, date_updated) VALUES
-- 	('5cf37266-3473-4006-984f-9325122678b7', 'true', 'bilibili', '21852', 'old-tomato', '[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv', 'flv', '', '', '', '2022-10-05 00:00:00', '2022-10-05 00:00:00')
-- 	ON CONFLICT DO NOTHING;

-- INSERT INTO shows (show_id, enable, platform, room_id, streamer_name, out_tmpl, parser, save_dir, post_cmds, split_rule, date_created, date_updated) VALUES
-- 	('5cf37266-3473-4006-984f-9325122678b8', 'false', 'bilibili', '21852', 'old-tomato', '[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv', 'flv', '', '', '', '2022-10-05 00:00:00', '2022-10-05 00:00:00')
-- 	ON CONFLICT DO NOTHING;

-- INSERT INTO shows (show_id, enable, platform, room_id, streamer_name, out_tmpl, parser, save_dir, post_cmds, split_rule, date_created, date_updated) VALUES
-- 	('5cf37266-3473-4006-984f-9325122678b9', 'false', 'bilibili', '21852', 'old-tomato', '[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv', 'flv', '', '', '', '2022-10-05 00:00:00', '2022-10-05 00:00:00')
-- 	ON CONFLICT DO NOTHING;

INSERT INTO configs (key, value, date_created, date_updated) VALUES
	('core_config', '{"PortalUsername":"olive","PortalPassword":"olive","LogDir":"/olive","SaveDir":"/olive","LogLevel":5,"SnapRestSeconds":15,"SplitRestSeconds":60,"CommanderPoolSize":1,"ParserMonitorRestSeconds":300,"DouyinCookie":"__ac_nonce=06245c89100e7ab2dd536; __ac_signature=_02B4Z6wo00f01LjBMSAAAIDBwA.aJ.c4z1C44TWAAEx696;","KuaishouCookie":"did=web_d86297aa2f579589b8abc2594b0ea985","BiliupEnable":true,"CookieFilepath":"","Threads":6}', '2022-10-05 00:00:00', '2022-10-05 00:00:00')
	ON CONFLICT DO NOTHING;