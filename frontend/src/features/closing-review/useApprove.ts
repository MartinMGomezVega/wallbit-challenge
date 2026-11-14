import { useMutation, useQueryClient } from '@tanstack/react-query';
import { approveClosing } from './api';
import { ApproveClosingResponse } from './schemas';

export interface UseApproveOptions {
  onSuccess?: (
    data: ApproveClosingResponse,
    variables: { closingId: string; version: number; clientNote?: string }
  ) => void;
  onError?: (error: unknown) => void;
}

export function useApprove(options?: UseApproveOptions) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      closingId,
      version,
      clientNote,
    }: {
      closingId: string;
      version: number;
      clientNote?: string;
    }) => approveClosing(closingId, version, clientNote),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['closings'] });
      queryClient.invalidateQueries({ queryKey: ['closings-summary'] });
      queryClient.invalidateQueries({ queryKey: ['closing-detail', variables.closingId] });
      options?.onSuccess?.(data, variables);
    },
    onError: (err) => {
      options?.onError?.(err);
    },
  });
}
