import { useMutation, useQueryClient } from '@tanstack/react-query';
import { importInvoicesAsAccountant } from '../api';
import { ImportResultDTO } from '../schemas';

export interface UseImportInvoicesOptions {
  onSuccess?: (data: ImportResultDTO) => void;
  onError?: (error: unknown) => void;
}

export function useImportInvoices(
  clientId: string,
  closingId: string,
  options?: UseImportInvoicesOptions
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (file: File) => importInvoicesAsAccountant(clientId, file),
    onSuccess: (data) => {
      if (data.imported > 0) {
        queryClient.invalidateQueries({ queryKey: ['closing-detail', closingId] });
        queryClient.invalidateQueries({ queryKey: ['closings'] });
        queryClient.invalidateQueries({ queryKey: ['closings-summary'] });
        queryClient.invalidateQueries({ queryKey: ['stats'] });
      }
      options?.onSuccess?.(data);
    },
    onError: (err) => {
      options?.onError?.(err);
    },
  });
}

