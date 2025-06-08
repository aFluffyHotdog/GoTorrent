TODO:
1. Write a proper test for the torrentFile Package
2. Rewrite the Client file to better support all the info protocol needs.
3. Support seeding
    The problem with that is because we are not sending out bitfield messages to anyone.
4. Support UDP torrents.

Currently struggling with sending handshakes out to peers. They all seem to time out :( womp womp

Update: It is struggling with making a connection in the first place, not sending over the handshakes