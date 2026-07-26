from __future__ import annotations

import copy
from collections.abc import Mapping
from dataclasses import dataclass
from typing import Any

from .image_artifact_private_storage import (
    ImageArtifactStoreResult,
    LocalPrivateImageArtifactStore,
)
from .image_backend_contract_fixture_client import (
    FAILURE_FIXTURE_DELIVERY_ALREADY_CONSUMED,
    FAILURE_FIXTURE_DELIVERY_MISMATCH,
    ContractFixtureImageBackendClient,
    FixtureBinaryDeliveryResult,
)
from .image_backend_profile_configuration import ImageBackendProfile
from .image_generation_adapter import (
    ImageGenerationAdapterResult,
    invoke_image_generation,
)


FAILURE_BINARY_DELIVERY_UNAVAILABLE = (
    "image_artifact_binary_delivery_unavailable"
)
FAILURE_BINARY_DELIVERY_MISMATCH = "image_artifact_binary_delivery_mismatch"
FAILURE_BINARY_DELIVERY_ALREADY_CONSUMED = (
    "image_artifact_binary_delivery_already_consumed"
)
FAILURE_PRIVATE_STORAGE = "image_artifact_private_storage_failed"

ZERO_EXTERNAL_SIDE_EFFECTS = {
    "artifact_upload_count": 0,
    "public_url_resolution_count": 0,
    "production_storage_write_count": 0,
    "retry_count": 0,
    "fallback_count": 0,
    "executor_call_count": 0,
    "confirmation_call_count": 0,
    "business_writeback_count": 0,
    "replay_call_count": 0,
}


@dataclass(frozen=True)
class ImageArtifactDeliveryCoordinatorResult:
    ok: bool
    citation: dict[str, Any] | None = None
    metadata_reference: dict[str, Any] | None = None
    failure_code: str | None = None
    failure_message: str = ""
    backend_call_count: int = 0
    image_generation_count: int = 0
    artifact_binary_delivery_count: int = 0
    local_artifact_store_write_count: int = 0
    artifact_binary_revalidation_count: int = 0


def invoke_and_store_fixture_image(
    intent_document: Mapping[str, Any],
    *,
    profile: ImageBackendProfile | None,
    client: ContractFixtureImageBackendClient,
    store: LocalPrivateImageArtifactStore,
    backend_request_id: str,
) -> ImageArtifactDeliveryCoordinatorResult:
    adapter_result = invoke_image_generation(
        intent_document,
        profile=profile,
        client=client,
        backend_request_id=backend_request_id,
    )
    if not adapter_result.ok:
        return failure_from_adapter(adapter_result)
    if (
        adapter_result.artifact_document is None
        or adapter_result.citation is None
        or adapter_result.metadata_reference is None
    ):
        return coordinator_failure(
            FAILURE_BINARY_DELIVERY_MISMATCH,
            "validated image artifact delivery state is incomplete",
            adapter_result=adapter_result,
        )

    store_result: ImageArtifactStoreResult | None = None

    def persist_fixture_payload(payload: bytes) -> None:
        nonlocal store_result
        try:
            candidate = store.put(adapter_result.artifact_document, payload)
        except Exception:
            return
        if isinstance(candidate, ImageArtifactStoreResult):
            store_result = candidate

    delivery_result = client.deliver_binary(
        adapter_result.artifact_document,
        persist_fixture_payload,
    )
    if not delivery_result.ok:
        return failure_from_delivery(adapter_result, delivery_result)
    if (
        store_result is None
        or not store_result.ok
        or store_result.stored_artifact is None
    ):
        return coordinator_failure(
            FAILURE_PRIVATE_STORAGE,
            "private image artifact storage failed",
            adapter_result=adapter_result,
            delivery_result=delivery_result,
            store_result=store_result,
        )

    return ImageArtifactDeliveryCoordinatorResult(
        ok=True,
        citation=copy.deepcopy(adapter_result.citation),
        metadata_reference=copy.deepcopy(adapter_result.metadata_reference),
        backend_call_count=adapter_result.backend_call_count,
        image_generation_count=adapter_result.image_generation_count,
        artifact_binary_delivery_count=(
            delivery_result.artifact_binary_delivery_count
        ),
        local_artifact_store_write_count=(
            store_result.local_artifact_store_write_count
        ),
        artifact_binary_revalidation_count=(
            store_result.artifact_binary_revalidation_count
        ),
    )


def coordinator_side_effect_counters(
    result: ImageArtifactDeliveryCoordinatorResult,
) -> dict[str, int]:
    return {
        "backend_call_count": result.backend_call_count,
        "image_generation_count": result.image_generation_count,
        "artifact_binary_delivery_count": (
            result.artifact_binary_delivery_count
        ),
        "local_artifact_store_write_count": (
            result.local_artifact_store_write_count
        ),
        "artifact_binary_revalidation_count": (
            result.artifact_binary_revalidation_count
        ),
        **ZERO_EXTERNAL_SIDE_EFFECTS,
    }


def failure_from_adapter(
    result: ImageGenerationAdapterResult,
) -> ImageArtifactDeliveryCoordinatorResult:
    return ImageArtifactDeliveryCoordinatorResult(
        ok=False,
        failure_code=result.failure_code,
        failure_message=result.failure_message,
        backend_call_count=result.backend_call_count,
        image_generation_count=result.image_generation_count,
    )


def failure_from_delivery(
    adapter_result: ImageGenerationAdapterResult,
    delivery_result: FixtureBinaryDeliveryResult,
) -> ImageArtifactDeliveryCoordinatorResult:
    if (
        delivery_result.failure_code
        == FAILURE_FIXTURE_DELIVERY_ALREADY_CONSUMED
    ):
        failure_code = FAILURE_BINARY_DELIVERY_ALREADY_CONSUMED
        failure_message = "fixture image artifact binary was already consumed"
    elif delivery_result.failure_code == FAILURE_FIXTURE_DELIVERY_MISMATCH:
        failure_code = FAILURE_BINARY_DELIVERY_MISMATCH
        failure_message = (
            "fixture image artifact binary did not match validated metadata"
        )
    else:
        failure_code = FAILURE_BINARY_DELIVERY_UNAVAILABLE
        failure_message = "fixture image artifact binary delivery failed"
    return coordinator_failure(
        failure_code,
        failure_message,
        adapter_result=adapter_result,
        delivery_result=delivery_result,
    )


def coordinator_failure(
    code: str,
    message: str,
    *,
    adapter_result: ImageGenerationAdapterResult,
    delivery_result: FixtureBinaryDeliveryResult | None = None,
    store_result: ImageArtifactStoreResult | None = None,
) -> ImageArtifactDeliveryCoordinatorResult:
    return ImageArtifactDeliveryCoordinatorResult(
        ok=False,
        failure_code=code,
        failure_message=message,
        backend_call_count=adapter_result.backend_call_count,
        image_generation_count=adapter_result.image_generation_count,
        artifact_binary_delivery_count=(
            delivery_result.artifact_binary_delivery_count
            if delivery_result is not None
            else 0
        ),
        local_artifact_store_write_count=(
            store_result.local_artifact_store_write_count
            if store_result is not None
            else 0
        ),
        artifact_binary_revalidation_count=(
            store_result.artifact_binary_revalidation_count
            if store_result is not None
            else 0
        ),
    )
