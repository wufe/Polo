import { ISession } from '@/state/models';
import React from 'react';
import { FramelessModal } from '../../../../modal/frameless-modal';
import { CommitMessage } from '../../../../shared/commit-message';

export const ApplicationSessionCommitModal = (props: { session: ISession; name: string; }) => {
    return <FramelessModal name={props.name}>
        <div className="overflow-hidden rounded-md text-nord0 dark:text-nord4">
            <CommitMessage {...props.session} />
        </div>
    </FramelessModal>
}