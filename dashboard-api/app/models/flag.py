from pydantic import BaseModel, Field


class FlagBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    enabled: bool = False
    description: str = Field(default="", max_length=500)


class FlagCreate(FlagBase):
    pass


class FlagUpdate(BaseModel):
    name: str | None = Field(default=None, min_length=1, max_length=100)
    enabled: bool | None = None
    description: str | None = Field(default=None, max_length=500)


class FlagResponse(FlagBase):
    id: str

    model_config = {"populate_by_name": True}
