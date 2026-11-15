import { apiFetch } from '../../shared/http/http';
import {
  ParametersDTO,
  ParametersDTOSchema,
  PutParametersResponse,
  PutParametersResponseSchema,
} from './schemas';

export async function getParameters(): Promise<ParametersDTO> {
  return apiFetch('/v1/parameters', ParametersDTOSchema);
}

export async function updateParameters(
  params: ParametersDTO
): Promise<PutParametersResponse> {
  return apiFetch('/v1/parameters', PutParametersResponseSchema, {
    method: 'PUT',
    body: JSON.stringify(params),
  });
}

