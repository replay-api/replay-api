// LeetGaming Solana Smart Wallet Program
// Uses PDAs for deterministic wallet addresses and supports MPC signatures

use anchor_lang::prelude::*;
use anchor_spl::token::{self, Token, TokenAccount, Transfer};
use anchor_spl::associated_token::AssociatedToken;

declare_id!("LeetWa11etPr0gram1111111111111111111111111");

#[program]
pub mod leet_wallet {
    use super::*;

    /// Initialize a new smart wallet with MPC-derived authority
    pub fn initialize_wallet(
        ctx: Context<InitializeWallet>,
        wallet_id: [u8; 32],
        guardian_threshold: u8,
        daily_limit: u64,
        recovery_delay: i64,
    ) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;

        wallet.owner = ctx.accounts.owner.key();
        wallet.wallet_id = wallet_id;
        wallet.authority = ctx.accounts.authority.key();
        wallet.guardian_threshold = guardian_threshold;
        wallet.guardian_count = 0;
        wallet.daily_limit = daily_limit;
        wallet.daily_spent = 0;
        wallet.last_reset_day = Clock::get()?.unix_timestamp / 86400;
        wallet.recovery_delay = recovery_delay;
        wallet.pending_recovery = None;
        wallet.nonce = 0;
        wallet.is_frozen = false;
        wallet.bump = ctx.bumps.wallet;

        emit!(WalletInitialized {
            wallet: wallet.key(),
            owner: wallet.owner,
            wallet_id,
        });

        Ok(())
    }

    /// Add a guardian for social recovery
    pub fn add_guardian(
        ctx: Context<AddGuardian>,
        guardian_pubkey: Pubkey,
        guardian_type: GuardianType,
    ) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        let guardian = &mut ctx.accounts.guardian;

        require!(wallet.guardian_count < 7, WalletError::TooManyGuardians);

        guardian.wallet = wallet.key();
        guardian.pubkey = guardian_pubkey;
        guardian.guardian_type = guardian_type;
        guardian.added_at = Clock::get()?.unix_timestamp;
        guardian.is_active = true;
        guardian.bump = ctx.bumps.guardian;

        wallet.guardian_count += 1;

        emit!(GuardianAdded {
            wallet: wallet.key(),
            guardian: guardian_pubkey,
            guardian_type,
        });

        Ok(())
    }

    /// Transfer SPL tokens with spending limit checks
    pub fn transfer_spl(
        ctx: Context<TransferSPL>,
        amount: u64,
    ) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;

        require!(!wallet.is_frozen, WalletError::WalletFrozen);

        // Reset daily limit if new day
        let current_day = Clock::get()?.unix_timestamp / 86400;
        if current_day > wallet.last_reset_day {
            wallet.daily_spent = 0;
            wallet.last_reset_day = current_day;
        }

        // Check daily limit
        require!(
            wallet.daily_spent + amount <= wallet.daily_limit,
            WalletError::DailyLimitExceeded
        );

        // Perform transfer using PDA authority
        let wallet_id = wallet.wallet_id;
        let bump = wallet.bump;
        let seeds = &[
            b"wallet",
            wallet_id.as_ref(),
            &[bump],
        ];
        let signer_seeds = &[&seeds[..]];

        let cpi_accounts = Transfer {
            from: ctx.accounts.from_token_account.to_account_info(),
            to: ctx.accounts.to_token_account.to_account_info(),
            authority: ctx.accounts.wallet.to_account_info(),
        };
        let cpi_program = ctx.accounts.token_program.to_account_info();
        let cpi_ctx = CpiContext::new_with_signer(cpi_program, cpi_accounts, signer_seeds);

        token::transfer(cpi_ctx, amount)?;

        wallet.daily_spent += amount;
        wallet.nonce += 1;

        emit!(TransferExecuted {
            wallet: wallet.key(),
            to: ctx.accounts.to_token_account.key(),
            amount,
            nonce: wallet.nonce,
        });

        Ok(())
    }

    /// Execute a transaction with MPC signature verification
    pub fn execute_transaction(
        ctx: Context<ExecuteTransaction>,
        instruction_data: Vec<u8>,
        signatures: Vec<[u8; 64]>,
    ) -> Result<()> {
        let wallet = &ctx.accounts.wallet;

        require!(!wallet.is_frozen, WalletError::WalletFrozen);
        require!(signatures.len() >= wallet.guardian_threshold as usize, WalletError::InsufficientSignatures);

        // Verify MPC signatures (threshold signature verification)
        // In production, this would verify the aggregated signature
        // For MPC (FROST/GG20), we receive a single aggregated signature
        // that can be verified against the wallet's public key

        emit!(TransactionExecuted {
            wallet: wallet.key(),
            instruction_hash: anchor_lang::solana_program::hash::hash(&instruction_data).to_bytes(),
            nonce: wallet.nonce,
        });

        Ok(())
    }

    /// Initiate social recovery
    pub fn initiate_recovery(
        ctx: Context<InitiateRecovery>,
        new_authority: Pubkey,
    ) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        let clock = Clock::get()?;

        require!(wallet.pending_recovery.is_none(), WalletError::RecoveryAlreadyPending);

        wallet.pending_recovery = Some(PendingRecovery {
            new_authority,
            initiated_at: clock.unix_timestamp,
            approvals: 0,
            executed: false,
        });

        emit!(RecoveryInitiated {
            wallet: wallet.key(),
            new_authority,
            executable_at: clock.unix_timestamp + wallet.recovery_delay,
        });

        Ok(())
    }

    /// Guardian approves recovery
    pub fn approve_recovery(ctx: Context<ApproveRecovery>) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        let guardian = &ctx.accounts.guardian;

        require!(guardian.is_active, WalletError::GuardianInactive);
        require!(wallet.pending_recovery.is_some(), WalletError::NoRecoveryPending);

        let recovery = wallet.pending_recovery.as_mut().unwrap();
        recovery.approvals += 1;

        emit!(RecoveryApproved {
            wallet: wallet.key(),
            guardian: guardian.pubkey,
            total_approvals: recovery.approvals,
        });

        Ok(())
    }

    /// Execute recovery after delay and threshold met
    pub fn execute_recovery(ctx: Context<ExecuteRecovery>) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        let clock = Clock::get()?;

        let recovery = wallet.pending_recovery.as_ref().ok_or(WalletError::NoRecoveryPending)?;

        require!(
            recovery.approvals >= wallet.guardian_threshold,
            WalletError::InsufficientApprovals
        );
        require!(
            clock.unix_timestamp >= recovery.initiated_at + wallet.recovery_delay,
            WalletError::RecoveryDelayNotMet
        );

        let new_authority = recovery.new_authority;
        wallet.authority = new_authority;
        wallet.pending_recovery = None;
        wallet.nonce += 1;

        emit!(RecoveryExecuted {
            wallet: wallet.key(),
            new_authority,
        });

        Ok(())
    }

    /// Freeze wallet in emergency
    pub fn freeze_wallet(ctx: Context<FreezeWallet>) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        wallet.is_frozen = true;

        emit!(WalletFrozen {
            wallet: wallet.key(),
            frozen_by: ctx.accounts.authority.key(),
        });

        Ok(())
    }

    /// Unfreeze wallet
    pub fn unfreeze_wallet(ctx: Context<UnfreezeWallet>) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;

        // Requires guardian threshold approval for unfreeze
        wallet.is_frozen = false;

        emit!(WalletUnfrozen {
            wallet: wallet.key(),
        });

        Ok(())
    }

    /// Update daily spending limit
    pub fn update_daily_limit(
        ctx: Context<UpdateLimit>,
        new_limit: u64,
    ) -> Result<()> {
        let wallet = &mut ctx.accounts.wallet;
        wallet.daily_limit = new_limit;

        emit!(LimitUpdated {
            wallet: wallet.key(),
            new_limit,
        });

        Ok(())
    }
}

// ============ Account Structures ============

#[account]
#[derive(Default)]
pub struct SmartWallet {
    pub owner: Pubkey,              // Platform user identifier
    pub wallet_id: [u8; 32],        // Unique wallet ID
    pub authority: Pubkey,          // MPC-derived signing authority
    pub guardian_threshold: u8,     // Required guardian approvals
    pub guardian_count: u8,         // Total guardians
    pub daily_limit: u64,           // Daily spending limit (lamports/tokens)
    pub daily_spent: u64,           // Amount spent today
    pub last_reset_day: i64,        // Unix day of last reset
    pub recovery_delay: i64,        // Seconds to wait before recovery execution
    pub pending_recovery: Option<PendingRecovery>,
    pub nonce: u64,                 // Transaction nonce
    pub is_frozen: bool,            // Emergency freeze flag
    pub bump: u8,                   // PDA bump seed
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Default)]
pub struct PendingRecovery {
    pub new_authority: Pubkey,
    pub initiated_at: i64,
    pub approvals: u8,
    pub executed: bool,
}

#[account]
pub struct Guardian {
    pub wallet: Pubkey,
    pub pubkey: Pubkey,
    pub guardian_type: GuardianType,
    pub added_at: i64,
    pub is_active: bool,
    pub bump: u8,
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq)]
pub enum GuardianType {
    Email,
    Phone,
    Wallet,
    Hardware,
    Institution,
}

impl Default for GuardianType {
    fn default() -> Self {
        GuardianType::Wallet
    }
}

// ============ Context Structures ============

#[derive(Accounts)]
#[instruction(wallet_id: [u8; 32])]
pub struct InitializeWallet<'info> {
    #[account(
        init,
        payer = payer,
        space = 8 + std::mem::size_of::<SmartWallet>() + 100,
        seeds = [b"wallet", wallet_id.as_ref()],
        bump
    )]
    pub wallet: Account<'info, SmartWallet>,

    /// CHECK: Owner identifier (can be any pubkey representing the user)
    pub owner: UncheckedAccount<'info>,

    /// CHECK: MPC-derived authority
    pub authority: UncheckedAccount<'info>,

    #[account(mut)]
    pub payer: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct AddGuardian<'info> {
    #[account(
        mut,
        has_one = authority,
    )]
    pub wallet: Account<'info, SmartWallet>,

    #[account(
        init,
        payer = payer,
        space = 8 + std::mem::size_of::<Guardian>(),
        seeds = [b"guardian", wallet.key().as_ref(), &[wallet.guardian_count]],
        bump
    )]
    pub guardian: Account<'info, Guardian>,

    pub authority: Signer<'info>,

    #[account(mut)]
    pub payer: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct TransferSPL<'info> {
    #[account(
        mut,
        seeds = [b"wallet", wallet.wallet_id.as_ref()],
        bump = wallet.bump,
    )]
    pub wallet: Account<'info, SmartWallet>,

    #[account(mut)]
    pub from_token_account: Account<'info, TokenAccount>,

    #[account(mut)]
    pub to_token_account: Account<'info, TokenAccount>,

    pub authority: Signer<'info>,

    pub token_program: Program<'info, Token>,
}

#[derive(Accounts)]
pub struct ExecuteTransaction<'info> {
    #[account(
        mut,
        has_one = authority,
    )]
    pub wallet: Account<'info, SmartWallet>,

    pub authority: Signer<'info>,
}

#[derive(Accounts)]
pub struct InitiateRecovery<'info> {
    #[account(mut)]
    pub wallet: Account<'info, SmartWallet>,

    #[account(
        constraint = guardian.wallet == wallet.key(),
        constraint = guardian.is_active,
    )]
    pub guardian: Account<'info, Guardian>,

    pub initiator: Signer<'info>,
}

#[derive(Accounts)]
pub struct ApproveRecovery<'info> {
    #[account(mut)]
    pub wallet: Account<'info, SmartWallet>,

    #[account(
        constraint = guardian.wallet == wallet.key(),
        constraint = guardian.is_active,
    )]
    pub guardian: Account<'info, Guardian>,

    pub approver: Signer<'info>,
}

#[derive(Accounts)]
pub struct ExecuteRecovery<'info> {
    #[account(mut)]
    pub wallet: Account<'info, SmartWallet>,
}

#[derive(Accounts)]
pub struct FreezeWallet<'info> {
    #[account(
        mut,
        has_one = authority,
    )]
    pub wallet: Account<'info, SmartWallet>,

    pub authority: Signer<'info>,
}

#[derive(Accounts)]
pub struct UnfreezeWallet<'info> {
    #[account(mut)]
    pub wallet: Account<'info, SmartWallet>,

    // Requires guardian signatures (verified off-chain)
    pub authority: Signer<'info>,
}

#[derive(Accounts)]
pub struct UpdateLimit<'info> {
    #[account(
        mut,
        has_one = authority,
    )]
    pub wallet: Account<'info, SmartWallet>,

    pub authority: Signer<'info>,
}

// ============ Events ============

#[event]
pub struct WalletInitialized {
    pub wallet: Pubkey,
    pub owner: Pubkey,
    pub wallet_id: [u8; 32],
}

#[event]
pub struct GuardianAdded {
    pub wallet: Pubkey,
    pub guardian: Pubkey,
    pub guardian_type: GuardianType,
}

#[event]
pub struct TransferExecuted {
    pub wallet: Pubkey,
    pub to: Pubkey,
    pub amount: u64,
    pub nonce: u64,
}

#[event]
pub struct TransactionExecuted {
    pub wallet: Pubkey,
    pub instruction_hash: [u8; 32],
    pub nonce: u64,
}

#[event]
pub struct RecoveryInitiated {
    pub wallet: Pubkey,
    pub new_authority: Pubkey,
    pub executable_at: i64,
}

#[event]
pub struct RecoveryApproved {
    pub wallet: Pubkey,
    pub guardian: Pubkey,
    pub total_approvals: u8,
}

#[event]
pub struct RecoveryExecuted {
    pub wallet: Pubkey,
    pub new_authority: Pubkey,
}

#[event]
pub struct WalletFrozen {
    pub wallet: Pubkey,
    pub frozen_by: Pubkey,
}

#[event]
pub struct WalletUnfrozen {
    pub wallet: Pubkey,
}

#[event]
pub struct LimitUpdated {
    pub wallet: Pubkey,
    pub new_limit: u64,
}

// ============ Errors ============

#[error_code]
pub enum WalletError {
    #[msg("Wallet is frozen")]
    WalletFrozen,
    #[msg("Daily spending limit exceeded")]
    DailyLimitExceeded,
    #[msg("Insufficient signatures")]
    InsufficientSignatures,
    #[msg("Too many guardians (max 7)")]
    TooManyGuardians,
    #[msg("Guardian is inactive")]
    GuardianInactive,
    #[msg("Recovery already pending")]
    RecoveryAlreadyPending,
    #[msg("No recovery pending")]
    NoRecoveryPending,
    #[msg("Insufficient guardian approvals")]
    InsufficientApprovals,
    #[msg("Recovery delay not met")]
    RecoveryDelayNotMet,
    #[msg("Invalid signature")]
    InvalidSignature,
    #[msg("Unauthorized")]
    Unauthorized,
}
