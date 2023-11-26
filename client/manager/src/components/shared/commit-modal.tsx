import { ISession } from '@polo/common/state/models';
import React from 'react';
import { DefaultModal } from '../modal/default-modal';
import { DefaultModalHeader, DefaultModalLayout } from '../modal/default-modal-layout/default-modal-layout';
import { FramelessModal } from '../modal/frameless-modal';
import { CommitMessage } from './commit-message';

type TProps = {
    name: string;
    title: string;
    commitMessage: string;
    commitAuthorName: string;
    commitAuthorEmail: string;
    commitDate: string;
}
export const CommitModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <DefaultModalLayout>
            <DefaultModalHeader>{props.title}</DefaultModalHeader>
            <div className="p-3">
                <CommitMessage {...props} />
            </div>
        </DefaultModalLayout>
    </DefaultModal>
}