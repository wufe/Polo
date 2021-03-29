import { ISession } from '@/state/models';
import React from 'react';
import { DefaultModal } from '../modal/default-modal';
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
        <div className="overflow-hidden lg:rounded-md text-nord0 dark:text-nord4 flex flex-col">
            <div className="mb-6">
                <div className="text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">{props.title}</div>
            </div>
            <CommitMessage {...props} />
        </div>
    </DefaultModal>
}