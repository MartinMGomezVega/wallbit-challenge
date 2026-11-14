import { apiFetch } from '../../shared/api/http';
import {
  ParametersDTO,
  ParametersDTOSchema,
  PutParametersResponse,
  PutParametersResponseSchema,
} from './schemas';

export async function getParameters(): Promise<ParametersDTO> {
  return apiFetch('/api/v1/accountant/parameters', ParametersDTOSchema);
}

export async function updateParameters(
  params: ParametersDTO
): Promise<PutParametersResponse> {
  return apiFetch('/api/v1/accountant/parameters', PutParametersResponseSchema, {
    method: 'PUT',
    body: JSON.stringify(params),
  });
}

