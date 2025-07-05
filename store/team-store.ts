import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  Team,
  TeamMember,
  TeamWithMembers,
  CreateTeamRequest,
  InviteTeamMemberRequest,
  ObjectId 
} from '@/types';

interface TeamState {
  // Current Team
  currentTeam: Team | null;
  
  // Teams
  teams: Team[];
  teamsLoading: boolean;
  
  // Members
  members: TeamMember[];
  membersLoading: boolean;
  
  // Invitations
  pendingInvitations: any[];
  invitationsLoading: boolean;
  
  // Loading
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadTeams: () => Promise<void>;
  loadTeam: (teamId: ObjectId) => Promise<void>;
  createTeam: (data: CreateTeamRequest) => Promise<Team>;
  updateTeam: (teamId: ObjectId, data: Partial<Team>) => Promise<void>;
  deleteTeam: (teamId: ObjectId) => Promise<void>;
  switchTeam: (teamId: ObjectId) => Promise<void>;
  
  // Members
  loadMembers: (teamId: ObjectId) => Promise<void>;
  inviteMember: (teamId: ObjectId, data: InviteTeamMemberRequest) => Promise<void>;
  updateMemberRole: (teamId: ObjectId, memberId: ObjectId, role: string) => Promise<void>;
  removeMember: (teamId: ObjectId, memberId: ObjectId) => Promise<void>;
  
  // Invitations
  loadInvitations: (teamId: ObjectId) => Promise<void>;
  cancelInvitation: (invitationId: ObjectId) => Promise<void>;
  resendInvitation: (invitationId: ObjectId) => Promise<void>;
  
  // Utility
  clearError: () => void;
}

export const useTeamStore = create<TeamState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      currentTeam: null,
      teams: [],
      teamsLoading: false,
      members: [],
      membersLoading: false,
      pendingInvitations: [],
      invitationsLoading: false,
      isLoading: false,
      error: null,

      // Load Teams
      loadTeams: async () => {
        set((state) => {
          state.teamsLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/teams');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load teams');
          }

          set((state) => {
            state.teams = result.data;
            state.teamsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load teams';
            state.teamsLoading = false;
          });
        }
      },

      // Load Team
      loadTeam: async (teamId: ObjectId) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/teams/${teamId}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load team');
          }

          set((state) => {
            state.currentTeam = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load team';
            state.isLoading = false;
          });
        }
      },

      // Create Team
      createTeam: async (data: CreateTeamRequest) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/teams', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to create team');
          }

          const newTeam: Team = result.data;

          set((state) => {
            state.teams.unshift(newTeam);
            state.currentTeam = newTeam;
            state.isLoading = false;
          });

          return newTeam;

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create team';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Update Team
      updateTeam: async (teamId: ObjectId, data: Partial<Team>) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/teams/${teamId}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to update team');
          }

          const updatedTeam: Team = result.data;

          set((state) => {
            const teamIndex = state.teams.findIndex(t => t._id === teamId);
            if (teamIndex !== -1) {
              state.teams[teamIndex] = updatedTeam;
            }
            if (state.currentTeam?._id === teamId) {
              state.currentTeam = updatedTeam;
            }
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to update team';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Delete Team
      deleteTeam: async (teamId: ObjectId) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/teams/${teamId}`, {
            method: 'DELETE',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to delete team');
          }

          set((state) => {
            state.teams = state.teams.filter(t => t._id !== teamId);
            if (state.currentTeam?._id === teamId) {
              state.currentTeam = null;
            }
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to delete team';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Switch Team
      switchTeam: async (teamId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/teams/${teamId}/switch`, {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to switch team');
          }

          await get().loadTeam(teamId);

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to switch team';
          });
          throw error;
        }
      },

      // Load Members
      loadMembers: async (teamId: ObjectId) => {
        set((state) => {
          state.membersLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/teams/${teamId}/members`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load members');
          }

          set((state) => {
            state.members = result.data;
            state.membersLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load members';
            state.membersLoading = false;
          });
        }
      },

      // Invite Member
      inviteMember: async (teamId: ObjectId, data: InviteTeamMemberRequest) => {
        try {
          const response = await fetch(`/api/client/teams/${teamId}/invite`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to invite member');
          }

          // Refresh invitations
          await get().loadInvitations(teamId);

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to invite member';
          });
          throw error;
        }
      },

      // Update Member Role
      updateMemberRole: async (teamId: ObjectId, memberId: ObjectId, role: string) => {
        try {
          const response = await fetch(`/api/client/teams/${teamId}/members/${memberId}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ role }),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to update member role');
          }

          set((state) => {
            const memberIndex = state.members.findIndex(m => m._id === memberId);
            if (memberIndex !== -1) {
              state.members[memberIndex].role = role as any;
            }
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to update member role';
          });
          throw error;
        }
      },

      // Remove Member
      removeMember: async (teamId: ObjectId, memberId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/teams/${teamId}/members/${memberId}`, {
            method: 'DELETE',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to remove member');
          }

          set((state) => {
            state.members = state.members.filter(m => m._id !== memberId);
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to remove member';
          });
          throw error;
        }
      },

      // Load Invitations
      loadInvitations: async (teamId: ObjectId) => {
        set((state) => {
          state.invitationsLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/teams/${teamId}/invitations`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load invitations');
          }

          set((state) => {
            state.pendingInvitations = result.data;
            state.invitationsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load invitations';
            state.invitationsLoading = false;
          });
        }
      },

      // Cancel Invitation
      cancelInvitation: async (invitationId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/invitations/${invitationId}`, {
            method: 'DELETE',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to cancel invitation');
          }

          set((state) => {
            state.pendingInvitations = state.pendingInvitations.filter(i => i._id !== invitationId);
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to cancel invitation';
          });
          throw error;
        }
      },

      // Resend Invitation
      resendInvitation: async (invitationId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/invitations/${invitationId}/resend`, {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to resend invitation');
          }

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to resend invitation';
          });
          throw error;
        }
      },

      // Clear Error
      clearError: () => {
        set((state) => {
          state.error = null;
        });
      },
    })),
    { name: 'team-store' }
  )
);
