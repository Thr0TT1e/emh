import { createClient, type Transport } from '@connectrpc/connect';

import { ApiKeyAdminService } from '~/sdk/emh/v1/api_key_admin_pb';
import { AuthService } from '~/sdk/emh/v1/auth_pb';
import { AwardAdminService } from '~/sdk/emh/v1/award_admin_pb';
import { AwardService } from '~/sdk/emh/v1/award_pb';
import { ConflictAdminService } from '~/sdk/emh/v1/conflict_admin_pb';
import { ConflictService } from '~/sdk/emh/v1/conflict_pb';
import { ContactService } from '~/sdk/emh/v1/contact_pb';
import { ExtractionService } from '~/sdk/emh/v1/extraction_pb';
import { HeroAdminService } from '~/sdk/emh/v1/hero_admin_pb';
import { HeroService } from '~/sdk/emh/v1/hero_pb';
import { LlmAdminService } from '~/sdk/emh/v1/llm_admin_pb';
import { LocationAdminService } from '~/sdk/emh/v1/location_admin_pb';
import { LocationService } from '~/sdk/emh/v1/location_pb';
import { MediaService } from '~/sdk/emh/v1/media_pb';
import { SubmissionAdminService, SubmissionService } from '~/sdk/emh/v1/submission_pb';

// Создаёт все Connect-клиенты поверх одного транспорта.
export function createApi(transport: Transport) {
  return {
    // Публичные
    hero: createClient(HeroService, transport),
    award: createClient(AwardService, transport),
    conflict: createClient(ConflictService, transport),
    location: createClient(LocationService, transport),
    submission: createClient(SubmissionService, transport),
    media: createClient(MediaService, transport),
    auth: createClient(AuthService, transport),
    contact: createClient(ContactService, transport),

    // Админские (требуют Authorization)
    heroAdmin: createClient(HeroAdminService, transport),
    awardAdmin: createClient(AwardAdminService, transport),
    conflictAdmin: createClient(ConflictAdminService, transport),
    locationAdmin: createClient(LocationAdminService, transport),
    apiKeyAdmin: createClient(ApiKeyAdminService, transport),
    submissionAdmin: createClient(SubmissionAdminService, transport),
    extraction: createClient(ExtractionService, transport),
    llmAdmin: createClient(LlmAdminService, transport),
  };
}

export type Api = ReturnType<typeof createApi>;
