db.createCollection('plans');

db.plans.createIndex({ 'baseentity._id': 1 }, { unique: true });

db.plans.insertMany([
  {
    baseentity: {
      _id: UUID('66bcfccc-2f92-4db9-a8dc-a0bcaaaf814c'),
      visibility_level: 4,
      visibility_type: 8,
      resource_owner: {
        tenant_id: UUID('a3a80810-f91c-4391-9eff-6d47a13bebde'),
        client_id: UUID('ff96c01f-a741-4429-a0cd-2868d408c42f'),
        group_id: UUID('1f743a05-03c7-471d-a6e9-de92d7097881'),
        user_id: UUID('c8e9301d-8d3d-4668-a7a1-a46fe813e9ed'),
      },
      created_at: new Date('2025-02-05T16:07:17.454+00:00'),
      updated_at: new Date('2025-02-05T16:07:17.454+00:00'),
    },
    name: 'Starter Pack',
    description: 'Kickstart your journey with essential features for free.',
    prices: {},
    operation_limits: [
      {
        operation_key: 'UploadReplayFiles',
        name: 'Upload Replay Files',
        description: 'Maximum number of replay files you can upload.',
        limit: 5,
      },
      {
        operation_key: 'StorageLimit',
        name: 'Storage Limit',
        description: 'Total storage space available for your files.',
        limit: 500, // in MB
      },
      {
        operation_key: 'FeatureAnalytics',
        name: 'Analytics',
        description: 'Access to basic analytics features.',
        limit: 1,
      },
      {
        operation_key: 'FeatureChatAndComments',
        name: 'Chat and Comments',
        description: 'Engage with the community through chat and comments.',
        limit: 1,
      },
      {
        operation_key: 'FeatureForum',
        name: 'Community Forum',
        description: 'Participate in community discussions through forums.',
        limit: 1,
      },
      {
        operation_key: 'FeatureAchievements',
        name: 'Achievements',
        description: 'Unlock and track your achievements.',
        limit: 1,
      },
    ],
    is_free: true,
  },
  {
    baseentity: {
      _id: UUID('66bcfccc-2f92-4db9-a8dc-a0bcaaaf814c'),
      visibility_level: 4,
      visibility_type: 8,
      resource_owner: {
        tenant_id: UUID('a3a80810-f91c-4391-9eff-6d47a13bebde'),
        client_id: UUID('ff96c01f-a741-4429-a0cd-2868d408c42f'),
        group_id: UUID('1f743a05-03c7-471d-a6e9-de92d7097881'),
        user_id: UUID('c8e9301d-8d3d-4668-a7a1-a46fe813e9ed'),
      },
      created_at: new Date('2025-02-05T16:07:17.454+00:00'),
      updated_at: new Date('2025-02-05T16:07:17.454+00:00'),
    },
    name: 'Pro Gamer Pack',
    description: 'Unlock advanced features and elevate your gaming experience.',
    prices: {
      monthly: [
        {
          amount: 4.99,
          currency: 'USD',
        },
      ],
      yearly: [
        {
          amount: 29.99, // 6 months free
          currency: 'USD',
        },
      ],
    },
    operation_limits: [
      {
        operation_key: 'UploadReplayFiles',
        name: 'Upload Replay Files',
        description: 'Maximum number of replay files you can upload.',
        limit: 100,
      },
      {
        operation_key: 'StorageLimit',
        name: 'Storage Limit',
        description: 'Total storage space available for your files.',
        limit: 10000, // in MB
      },
      {
        operation_key: 'FeatureAnalytics',
        name: 'Advanced Analytics',
        description: 'Gain insights with advanced analytics features.',
        limit: 1,
      },
      {
        operation_key: 'FeatureChatAndComments',
        name: 'Enhanced Chat and Comments',
        description: 'Engage deeply with the community through enhanced chat and comments.',
        limit: 1,
      },
      {
        operation_key: 'FeatureForum',
        name: 'Community Forum',
        description: 'Participate in community discussions through forums.',
        limit: 1,
      },
      {
        operation_key: 'FeatureAchievements',
        name: 'Achievements',
        description: 'Unlock and track your achievements.',
        limit: 1,
      },
      {
        operation_key: 'MediaAlbum',
        name: 'Media Album',
        description: 'Create and manage your media albums.',
        limit: 1,
      },
      {
        operation_key: 'MintAssetAmount',
        name: 'Mint Assets',
        description: 'Mint new assets in the Web 3.0 ecosystem.',
        limit: 10,
      },
      {
        operation_key: 'ListAssetAmount',
        name: 'List Assets',
        description: 'List your assets for trading or sale.',
        limit: 10,
      },
      {
        operation_key: 'SwapAssetAmount',
        name: 'Swap Assets',
        description: 'Swap assets with other users.',
        limit: 10,
      },
      {
        operation_key: 'StakeAssetAmount',
        name: 'Stake Assets',
        description: 'Stake your assets to earn rewards.',
        limit: 10,
      },
      {
        operation_key: 'SquadProfileAmount',
        name: 'Squad Profiles',
        description: 'Create and manage squad profiles.',
        limit: 5,
      },
      {
        operation_key: 'PlayersPerSquadAmount',
        name: 'Players per Squad',
        description: 'Maximum number of players per squad.',
        limit: 12,
      },
      {
        operation_key: 'PlayerProfileAmount',
        name: 'Player Profiles',
        description: 'Create and manage player profiles.',
        limit: 12,
      },
      {
        operation_key: 'SquadProfileBoostAmount',
        name: 'Squad Profile Boosts',
        description: 'Boost your squad profiles for better visibility.',
        limit: 5,
      },
      {
        operation_key: 'MatchMakingQueueAmount',
        name: 'Matchmaking Queues',
        description: 'Create and manage matchmaking queues.',
        limit: 500, // monthly
      },
      {
        operation_key: 'MatchMakingCreateAmount',
        name: 'Create Matchmaking',
        description: 'Create matchmaking events.',
        limit: 500, // monthly
      },
      {
        operation_key: 'MatchMakingPriorityQueue',
        name: 'Priority Matchmaking Queue',
        description: 'Get priority access to matchmaking queues.',
        limit: 1,
      },
      {
        operation_key: 'MatchMakingCreateFeaturedBoosts',
        name: 'Featured Matchmaking Boosts',
        description: 'Boost your matchmaking events for better visibility.',
        limit: 5,
      },
      {
        operation_key: 'StoreAmount',
        name: 'Store',
        description: 'Create and manage your store.',
        limit: 1,
      },
      {
        operation_key: 'StoreBoostAmount',
        name: 'Store Boosts',
        description: 'Boost your store for better visibility.',
        limit: 5,
      },
      {
        operation_key: 'StoreVoucherAmount',
        name: 'Store Vouchers',
        description: 'Create and manage store vouchers.',
        limit: 50,
      },
      {
        operation_key: 'StorePromotionAmount',
        name: 'Store Promotions',
        description: 'Create and manage store promotions.',
        limit: 5,
      },
      {
        operation_key: 'FeatureStoreCatalogIntegration',
        name: 'Store Catalog Integration',
        description: 'Integrate your store with external catalogs.',
        limit: 1,
      },
      {
        operation_key: 'FeatureGlobalWallet',
        name: 'Global Wallet',
        description: 'Access to a global wallet for transactions.',
        limit: 1,
      },
      {
        operation_key: 'GlobalWalletLimit',
        name: 'Global Wallet Limit',
        description: 'Maximum limit for your global wallet.',
        limit: 10000, // in currency units
      },
      {
        operation_key: 'GlobalPaymentProcessingAmount',
        name: 'Global Payment Processing',
        description: 'Process payments globally.',
        limit: 100,
      },
      {
        operation_key: 'FeatureBetting',
        name: 'Betting',
        description: 'Access to betting features.',
        limit: 1,
      },
      {
        operation_key: 'Notifications',
        name: 'Notifications',
        description: 'Receive notifications for various events.',
        limit: 1,
      },
      {
        operation_key: 'CustomEmails',
        name: 'Custom Emails',
        description: 'Send custom emails to users.',
        limit: 100,
      },
      {
        operation_key: 'PremiumSupport',
        name: 'Premium Support',
        description: 'Access to premium support services.',
        limit: 1,
      },
      {
        operation_key: 'DiscordIntegration',
        name: 'Discord Integration',
        description: 'Integrate with Discord for community engagement.',
        limit: 1,
      },
      {
        operation_key: 'FeatureAPI',
        name: 'API Access',
        description: 'Access to API for integrations.',
        limit: 1,
      },
      {
        operation_key: 'FeatureWebhooks',
        name: 'Webhooks',
        description: 'Set up webhooks for real-time notifications.',
        limit: 1,
      },
      {
        operation_key: 'ExclusiveContent',
        name: 'Exclusive Content',
        description: 'Access to exclusive content.',
        limit: 1,
      },
      {
        operation_key: 'AdFreeExperience',
        name: 'Ad-Free Experience',
        description: 'Enjoy an ad-free experience.',
        limit: 1,
      },
      {
        operation_key: 'EarlyAccessFeatures',
        name: 'Early Access to Features',
        description: 'Get early access to new features.',
        limit: 1,
      },
    ],
    is_free: false,
  },
  {
    baseentity: {
      _id: UUID('77bcfccc-3f92-4db9-a8dc-a0bcaaaf814d'),
      visibility_level: 5,
      visibility_type: 9,
      resource_owner: {
        tenant_id: UUID('b3b80810-f91c-4391-9eff-6d47a13bebdf'),
        client_id: UUID('gg96c01f-a741-4429-a0cd-2868d408c42g'),
        group_id: UUID('2f743a05-03c7-471d-a6e9-de92d7097882'),
        user_id: UUID('d8e9301d-8d3d-4668-a7a1-a46fe813e9ee'),
      },
      created_at: new Date('2025-02-05T16:07:17.454+00:00'),
      updated_at: new Date('2025-02-05T16:07:17.454+00:00'),
    },
    name: 'Ultimate Gamer Pack',
    description: 'Experience the ultimate gaming journey with exclusive features and benefits.',
    prices: {
      monthly: [
        {
          amount: 9.99,
          currency: 'USD',
        },
      ],
      yearly: [
        {
          amount: 49.99, // 6 months free
          currency: 'USD',
        },
      ],
    },
    operation_limits: [
      {
        operation_key: 'UploadReplayFiles',
        name: 'Upload Replay Files',
        description: 'Unlimited number of replay files you can upload.',
        limit: -1,
      },
      {
        operation_key: 'StorageLimit',
        name: 'Storage Limit',
        description: 'Massive storage space available for your files.',
        limit: 50000, // in MB
      },
      {
        operation_key: 'FeatureAnalytics',
        name: 'Pro Analytics',
        description: 'Gain deep insights with professional analytics features.',
        limit: 1,
      },
      {
        operation_key: 'FeatureChatAndComments',
        name: 'Premium Chat and Comments',
        description: 'Engage with the community through premium chat and comments.',
        limit: 1,
      },
      {
        operation_key: 'FeatureForum',
        name: 'Exclusive Community Forum',
        description: 'Participate in exclusive community discussions through forums.',
        limit: 1,
      },
      {
        operation_key: 'FeatureAchievements',
        name: 'Elite Achievements',
        description: 'Unlock and track your elite achievements.',
        limit: 1,
      },
      {
        operation_key: 'MediaAlbum',
        name: 'Media Album',
        description: 'Create and manage your media albums.',
        limit: -1,
      },
      {
        operation_key: 'MintAssetAmount',
        name: 'Mint Assets',
        description: 'Mint new assets in the Web 3.0 ecosystem.',
        limit: 20,
      },
      {
        operation_key: 'ListAssetAmount',
        name: 'List Assets',
        description: 'List your assets for trading or sale.',
        limit: 20,
      },
      {
        operation_key: 'SwapAssetAmount',
        name: 'Swap Assets',
        description: 'Swap assets with other users.',
        limit: 20,
      },
      {
        operation_key: 'StakeAssetAmount',
        name: 'Stake Assets',
        description: 'Stake your assets to earn rewards.',
        limit: 20,
      },
      {
        operation_key: 'SquadProfileAmount',
        name: 'Squad Profiles',
        description: 'Create and manage squad profiles.',
        limit: 10,
      },
      {
        operation_key: 'PlayersPerSquadAmount',
        name: 'Players per Squad',
        description: 'Maximum number of players per squad.',
        limit: 50,
      },
      {
        operation_key: 'PlayerProfileAmount',
        name: 'Player Profiles',
        description: 'Create and manage player profiles.',
        limit: 100,
      },
      {
        operation_key: 'SquadProfileBoostAmount',
        name: 'Squad Profile Boosts',
        description: 'Boost your squad profiles for better visibility.',
        limit: 10,
      },
      {
        operation_key: 'MatchMakingQueueAmount',
        name: 'Matchmaking Queues',
        description: 'Create and manage matchmaking queues.',
        limit: -1,
      },
      {
        operation_key: 'MatchMakingCreateAmount',
        name: 'Create Matchmaking',
        description: 'Create matchmaking events.',
        limit: -1,
      },
      {
        operation_key: 'MatchMakingPriorityQueue',
        name: 'Priority Matchmaking Queue',
        description: 'Get priority access to matchmaking queues.',
        limit: 2,
      },
      {
        operation_key: 'MatchMakingCreateFeaturedBoosts',
        name: 'Featured Matchmaking Boosts',
        description: 'Boost your matchmaking events for better visibility.',
        limit: 10,
      },
      {
        operation_key: 'StoreAmount',
        name: 'Store',
        description: 'Create and manage your store.',
        limit: 1,
      },
      {
        operation_key: 'StoreBoostAmount',
        name: 'Store Boosts',
        description: 'Boost your store for better visibility.',
        limit: 10,
      },
      {
        operation_key: 'StoreVoucherAmount',
        name: 'Store Vouchers',
        description: 'Create and manage store vouchers.',
        limit: 20,
      },
      {
        operation_key: 'StorePromotionAmount',
        name: 'Store Promotions',
        description: 'Create and manage store promotions.',
        limit: 10,
      },
      {
        operation_key: 'FeatureStoreCatalogIntegration',
        name: 'Store Catalog Integration',
        description: 'Integrate your store with external catalogs.',
        limit: 1,
      },
      {
        operation_key: 'FeatureGlobalWallet',
        name: 'Global Wallet',
        description: 'Access to a global wallet for transactions.',
        limit: 1,
      },
      {
        operation_key: 'GlobalWalletLimit',
        name: 'Global Wallet Limit',
        description: 'Maximum limit for your global wallet.',
        limit: 50000, // in currency units
      },
      {
        operation_key: 'GlobalPaymentProcessingAmount',
        name: 'Global Payment Processing',
        description: 'Process payments globally.',
        limit: 500,
      },
      {
        operation_key: 'FeatureBetting',
        name: 'Betting',
        description: 'Access to betting features.',
        limit: 1,
      },
      {
        operation_key: 'Notifications',
        name: 'Notifications',
        description: 'Receive notifications for various events.',
        limit: 1,
      },
      {
        operation_key: 'CustomEmails',
        name: 'Custom Emails',
        description: 'Send custom emails to users.',
        limit: 500,
      },
      {
        operation_key: 'PremiumSupport',
        name: 'Premium Support',
        description: 'Access to premium support services.',
        limit: 1,
      },
      {
        operation_key: 'DiscordIntegration',
        name: 'Discord Integration',
        description: 'Integrate with Discord for community engagement.',
        limit: 1,
      },
      {
        operation_key: 'FeatureAPI',
        name: 'API Access',
        description: 'Access to API for integrations.',
        limit: 1,
      },
      {
        operation_key: 'FeatureWebhooks',
        name: 'Webhooks',
        description: 'Set up webhooks for real-time notifications.',
        limit: 1,
      },
      {
        operation_key: 'ExclusiveContent',
        name: 'Exclusive Content',
        description: 'Access to exclusive content.',
        limit: 1,
      },
      {
        operation_key: 'AdFreeExperience',
        name: 'Ad-Free Experience',
        description: 'Enjoy an ad-free experience.',
        limit: 1,
      },
      {
        operation_key: 'EarlyAccessFeatures',
        name: 'Early Access to Features',
        description: 'Get early access to new features.',
        limit: 1,
      },
    ],
    is_free: false,
  },
]);