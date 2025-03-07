package gov

import (
	"encoding/hex"
	"fmt"

	modulestypes "github.com/forbole/bdjuno/v2/modules/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	parsecmdtypes "github.com/forbole/juno/v3/cmd/parse/types"
	"github.com/forbole/juno/v3/types/config"
	"github.com/spf13/cobra"

	"github.com/forbole/juno/v3/parser"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/gov"
	"github.com/forbole/bdjuno/v2/utils"
)

// proposalCmd returns the Cobra command allowing to fix all things related to a proposal
func proposalCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "proposal [id]",
		Short: "Get the description, votes and everything related to a proposal given its id",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID := args[0]

			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			sources, err := modulestypes.BuildSources(config.Cfg.Node, parseCtx.EncodingConfig)
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)

			// Build the gov module
			govModule := gov.NewModule(sources.GovSource, nil, nil, nil, nil, nil, parseCtx.EncodingConfig.Marshaler, db)

			err = refreshProposalDetails(parseCtx, proposalID, govModule)
			if err != nil {
				return err
			}

			err = refreshProposalDeposits(parseCtx, proposalID, govModule)
			if err != nil {
				return err
			}

			err = refreshProposalVotes(parseCtx, proposalID, govModule)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func refreshProposalDetails(parseCtx *parser.Context, proposalID string, govModule *gov.Module) error {
	// Get the tx that created the proposal
	txs, err := utils.QueryTxs(parseCtx.Node, fmt.Sprintf("submit_proposal.proposal_id=%s", proposalID))
	if err != nil {
		return err
	}

	if len(txs) > 1 {
		return fmt.Errorf("expecting only one create proposal transaction, found %d", len(txs))
	}

	// Get the tx details
	tx, err := parseCtx.Node.Tx(hex.EncodeToString(txs[0].Tx.Hash()))
	if err != nil {
		return err
	}

	// Handle the MsgSubmitProposal messages
	for index, msg := range tx.GetMsgs() {
		if _, ok := msg.(*govtypes.MsgSubmitProposal); !ok {
			continue
		}

		err = govModule.HandleMsg(index, msg, tx)
		if err != nil {
			return fmt.Errorf("error while handling MsgSubmitProposal: %s", err)
		}
	}

	return nil
}

func refreshProposalDeposits(parseCtx *parser.Context, proposalID string, govModule *gov.Module) error {
	// Get the tx that deposited to the proposal
	txs, err := utils.QueryTxs(parseCtx.Node, fmt.Sprintf("proposal_deposit.proposal_id=%s", proposalID))
	if err != nil {
		return err
	}

	for _, tx := range txs {
		// Get the tx details
		junoTx, err := parseCtx.Node.Tx(hex.EncodeToString(tx.Tx.Hash()))
		if err != nil {
			return err
		}

		// Handle the MsgDeposit messages
		for index, msg := range junoTx.GetMsgs() {
			if _, ok := msg.(*govtypes.MsgDeposit); !ok {
				continue
			}

			err = govModule.HandleMsg(index, msg, junoTx)
			if err != nil {
				return fmt.Errorf("error while handling MsgDeposit: %s", err)
			}
		}
	}

	return nil
}

func refreshProposalVotes(parseCtx *parser.Context, proposalID string, govModule *gov.Module) error {
	// Get the tx that voted the proposal
	txs, err := utils.QueryTxs(parseCtx.Node, fmt.Sprintf("proposal_vote.proposal_id=%s", proposalID))
	if err != nil {
		return err
	}

	for _, tx := range txs {
		// Get the tx details
		junoTx, err := parseCtx.Node.Tx(hex.EncodeToString(tx.Tx.Hash()))
		if err != nil {
			return err
		}

		// Handle the MsgVote messages
		for index, msg := range junoTx.GetMsgs() {
			if _, ok := msg.(*govtypes.MsgVote); !ok {
				continue
			}

			err = govModule.HandleMsg(index, msg, junoTx)
			if err != nil {
				return fmt.Errorf("error while handling MsgVote: %s", err)
			}
		}
	}

	return nil
}
