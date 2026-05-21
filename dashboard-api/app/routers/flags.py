from bson import ObjectId
from fastapi import APIRouter, HTTPException, status

from app.db import get_database
from app.models.flag import FlagCreate, FlagResponse, FlagUpdate

router = APIRouter(prefix="/flags", tags=["flags"])


def _serialize(doc: dict) -> FlagResponse:
    return FlagResponse(
        id=str(doc["_id"]),
        name=doc["name"],
        enabled=doc["enabled"],
        description=doc.get("description", ""),
    )


def _valid_object_id(flag_id: str) -> ObjectId:
    if not ObjectId.is_valid(flag_id):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid flag ID"
        )
    return ObjectId(flag_id)


@router.get("/", response_model=list[FlagResponse])
async def list_flags() -> list[FlagResponse]:
    db = get_database()
    return [_serialize(doc) async for doc in db["flags"].find()]


@router.post("/", response_model=FlagResponse, status_code=status.HTTP_201_CREATED)
async def create_flag(payload: FlagCreate) -> FlagResponse:
    db = get_database()
    result = await db["flags"].insert_one(payload.model_dump())
    doc = await db["flags"].find_one({"_id": result.inserted_id})
    return _serialize(doc)


@router.get("/{flag_id}", response_model=FlagResponse)
async def get_flag(flag_id: str) -> FlagResponse:
    oid = _valid_object_id(flag_id)
    db = get_database()
    doc = await db["flags"].find_one({"_id": oid})
    if not doc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Flag not found")
    return _serialize(doc)


@router.patch("/{flag_id}", response_model=FlagResponse)
async def update_flag(flag_id: str, payload: FlagUpdate) -> FlagResponse:
    oid = _valid_object_id(flag_id)
    update_data = {k: v for k, v in payload.model_dump().items() if v is not None}
    if not update_data:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail="No fields to update"
        )
    db = get_database()
    await db["flags"].update_one({"_id": oid}, {"$set": update_data})
    doc = await db["flags"].find_one({"_id": oid})
    if not doc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Flag not found")
    return _serialize(doc)


@router.delete("/{flag_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_flag(flag_id: str) -> None:
    oid = _valid_object_id(flag_id)
    db = get_database()
    result = await db["flags"].delete_one({"_id": oid})
    if result.deleted_count == 0:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Flag not found")
