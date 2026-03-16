-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN public_tracker_list TEXT NOT NULL DEFAULT '1337x,rarbg,yts,nyaa,eztv,limetorrents,thepiratebay,kickasstorrents,torrentz2,glodls,magnetdl,ettv,isohunt,bt4g,solidtorrents,bitsearch,torrentgalaxy,fitgirl';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN public_tracker_list;
