export type EvaluationResponse = {
  success: true;
  result: "on" | "off";
};

export type AuthCheckResponse = {
  success: true;
  data: {
    tenant_id: number;
    app_id: string;
    subject: string;
  };
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

