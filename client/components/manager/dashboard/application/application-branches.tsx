import { IApplicationBranchModel } from '@/state/models';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React from 'react';
import dayjs from 'dayjs';

type TProps = {
    branches: IApplicationBranchModel[];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ApplicationBranches = observer((props: TProps) => {
    return <>
        <h4 className="my-2 text-gray-700 dark:text-gray-300 uppercase">Branches:</h4>
        <div className="divide-y dark:divide-gray-700">
            {sortBranches(props.branches).map((branch, key) =>
                <div
                    className="grid items-start gap-2 grid-cols-12 pt-6 pb-4">
                    <React.Fragment key={key}>
                        <span className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis col-span-5" title={branch.name}>{branch.name}</span>
                        <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis col-span-5">{branch.message}</div>
                        <span className=" text-center col-span-2 row-span-2">
                            <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                        </span>
                        <div className="col-span-5">
                            <div className="text-xs text-gray-500 uppercase">Last author</div>
                            <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.author}</div>
                        </div>
                        <div className="col-span-5">
                            <div className="text-xs text-gray-500 uppercase">Updated</div>
                            <div className="text-sm whitespace-nowrap">{dayjs(branch.date).format('DD MMM HH:mm')}</div>
                        </div>
                    </React.Fragment>
                </div>)}
        </div>
        
    </>;
});

function sortBranches(branches: IApplicationBranchModel[]): IApplicationBranchModel[] {
    const preferred = [
        'master',
        'main',
        'hotfix',
        'develop',
        'dev',
        'feature'
    ];

    let result: IApplicationBranchModel[] = [];
    let length = branches.length;

    for (const pref of preferred) {
        for (let i = 0; i < length; i++) {
            const branch = branches[i];
            if (branch.name.toLowerCase().startsWith(pref.toLowerCase())) {
                result.push(branch);
                branches.splice(i, 1)
                length--;
                i--;
            }
        }
    }

    return result.concat(branches);
}