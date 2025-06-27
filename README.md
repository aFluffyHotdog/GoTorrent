TODO:
1. Write a proper test for the torrentFile Package
2. Rewrite the Client file to better support all the info protocol needs.
3. Support seeding
    The problem with that is because we are not sending out bitfield messages to anyone.
4. Support multiple file torrents?

24/6: Right now the program terminates after finishing the handshakes, will have to see what's wrong. Turns out the pieceHashes is empty, need to check how it's being parsed




Currently struggling with sending handshakes out to peers. They all seem to time out :( womp womp

Error 3: Woops I added the wrong number of bytes. But now the client won't read anything after sending a handshake
As a sanity check, it is saving the right IP address and infohash
Update: It seems to be sending the same peerID in every handshake???? No... that's supposed to happen
Plot twist: It was working fine.... you forgot to print out that these clients were successful in connecting



Error 2: Looks like we're getting a slice out of bound error when we're trying to serialize a message

Error 1 Update: 
It is struggling with making a connection in the first place, not sending over the handshakes
Turns out the peers decoded from the tracker response is not a divisor of 6, something is up. I should've used a string instead of []byte