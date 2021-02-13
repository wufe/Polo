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
        <h4 className="my-2 text-gray-500 text-xs uppercase">Branches:</h4>
        <div className="divide-y dark:divide-gray-700">
            {sortBranches(props.branches).map((branch, key) =>
                <div
                    className="grid items-start gap-2 grid-cols-12 pt-6 pb-4" key={key}>
                    <div className="col-span-10">
                        <div className="text-xs text-gray-500 uppercase">Name</div>
                        <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis" title={branch.name}>{branch.name}</div>
                    </div>
                    <span className="text-center col-span-2">
                        <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                    </span>
                    <div className="col-span-12">
                        <div className="grid grid-cols-12 gap-2 items-center p-3 bg-nord5 dark:bg-nord1 rounded-sm">
                            <div className="col-span-12">
                                <div className="text-xs text-gray-500 uppercase">Message</div>
                                <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.message}</div>
                            </div>
                            <div className="col-span-5">
                                <div className="text-xs text-gray-500 uppercase">Author</div>
                                <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.author}</div>
                            </div>
                            <div className="col-span-7">
                                <div className="text-xs text-gray-500 uppercase">Date</div>
                                <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{dayjs(branch.date).format('DD MMM HH:mm')}</div>
                            </div>
                        </div>
                        
                    </div>
                    
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