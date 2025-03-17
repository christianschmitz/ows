package main

import (
	//"log"

	//"ows/ledger"
	//"ows/network"
)

//func getKeyPair() *ledger.KeyPair {
//	p, err := ledger.ReadKeyPair(ledger.HomeDir+"/key", true)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return p
//}

//func getSyncedLedgerClient() *sync.LedgerClient {
//	l, err := ledger.ReadLedger(false)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	c := sync.NewLedgerClient(l)
//
//	if err := c.Sync(); err != nil {
//		log.Fatal(err)
//	}
//
//	return c
//}
//
//func readLedger() *ledger.Ledger {
//	c := getSyncedLedgerClient()
//
//	return c.Ledger
//}
//
//func signAndSubmitChangeSet(client *sync.LedgerClient, cs *ledger.ChangeSet) error {
//	key := getKeyPair()
//
//	signature, err := key.SignChangeSet(cs)
//	if err != nil {
//		return err
//	}
//
//	cs.Signatures = append(cs.Signatures, signature)
//
//	if err := client.PublishChangeSet(cs); err != nil {
//		return err
//	}
//
//	if err := client.Ledger.AppendChangeSet(cs, false); err != nil {
//		return err
//	}
//
//	client.Ledger.Write()
//
//	return err
//}
//
//// creates change set, signs it, then submits it
//func createChangeSet(actions ...ledger.Action) {
//	c := getSyncedLedgerClient()
//	cs := c.Ledger.NewChangeSet(actions...)
//	if err := signAndSubmitChangeSet(c, cs); err != nil {
//		log.Fatal(err)
//	}
//}
//