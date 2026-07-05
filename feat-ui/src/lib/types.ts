export type Flag = {
  id: number;
  tenant_id: number;
  key: string;
  description: string;
  default_enabled: boolean;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
};

export type FlagListResponse = {
  success: true;
  data: {
    flags: Flag[];
  };
};

export type FlagResponse = {
  success: true;
  data: Flag;
};

export type BulkSetResponse = {
  success: true;
  data: {
    applied: number;
  };
};

export type EvaluationResponse = {
  success: true;
  result: "on" | "off";
};

export type ErrorResponse = {
  success: false;
  error: {
    code: string;
    message: string;
    details?: Array<{
      field?: string;
      message: string;
    }>;
  };
};

export type FlagDraft = {
  key: string;
  description: string;
  default_enabled: boolean;
};

