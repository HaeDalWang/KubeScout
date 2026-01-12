export type DriftStatus = "SYNC" | "PATCH_DRIFT" | "MINOR_DRIFT" | "MAJOR_DRIFT" | "UNKNOWN";

export interface Release {
    name: string;
    namespace: string;
    chart_name: string;
    chart_version: string;
    app_version: string;
    revision: number;
    updated: string;
    icon?: string;
}

export interface ComparisonResult {
    release: Release;
    latest_version: string;
    latest_app_version: string;
    status: DriftStatus;
    upstream_url: string;
    checked_at: string;
}
