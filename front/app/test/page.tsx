import { TestComboBox } from "@/components/custom/test/test-combobox";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { DatePicker } from "@/components/ui/date-picker";
import {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogContent,
  ResponsiveDialogDescription,
  ResponsiveDialogFooter,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
  ResponsiveDialogTrigger,
} from "@/components/custom/responsive-dialog";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import TestToast from "@/components/custom/test/test-toast";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { AppLayout } from "../layout";
import { ComponentExample } from "@/components/component-example";

export default function testPage() {
  return (
    <AppLayout>
    <div className="grid grid-cols-3 gap-4 p-4">
      {/* Alert */}
      <Card>
        <CardHeader>
          <CardTitle>Alert</CardTitle>
          <CardDescription>Card Description</CardDescription>
          <CardAction>Card Action</CardAction>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertTitle>This is a test page</AlertTitle>
            <AlertDescription>
              This page includes various UI components for testing purposes.
            </AlertDescription>
          </Alert>
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>
      {/* Badges */}
      <Card>
        <CardHeader>
          <CardTitle>Badges</CardTitle>
          <CardDescription>With more content</CardDescription>
        </CardHeader>
        <CardContent>
          <Badge variant="default">New</Badge>
          <Badge variant="destructive" className="ml-2">
            Error
          </Badge>
          <Badge variant="outline" className="ml-2">
            Outline
          </Badge>
          <Badge variant="secondary" className="ml-2">
            Secondary
          </Badge>
          <Badge variant="link" className="ml-2">
            Success
          </Badge>
        </CardContent>
        <CardFooter>
          <p>Footer Content</p>
        </CardFooter>
      </Card>
      {/* Buttons */}
      <Card>
        <CardHeader>
          <CardTitle>Buttons</CardTitle>
          <CardDescription>With more content</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="default">Default</Button>
          <Button variant="destructive" className="ml-2">
            Destructive
          </Button>
          <Button variant="outline" className="ml-2">
            Outline
          </Button>
          <Button variant="secondary" className="ml-2">
            Secondary
          </Button>
          <Button variant="link" className="ml-2">
            Link
          </Button>
        </CardContent>
        <CardFooter>
          <p>Footer Content</p>
        </CardFooter>
      </Card>
      {/* Form Fields, Inputs, Combobox */}
      <Card>
        <CardHeader>
          <CardTitle>Form Fields, Inputs, Combobox</CardTitle>
          <CardDescription>With more content</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-2">
          <Field>
            <Label>Enter something:</Label>
            <Input placeholder="Type something..."  />
          </Field>
          <Field>
            <Label>Select a framework:</Label>
            <TestComboBox />
          </Field>
          <Field>
            <Label>Another Input:</Label>
            <DatePicker />
          </Field>
        </CardContent>
        <CardFooter>
          <p>Footer Content</p>
        </CardFooter>
      </Card>
      {/* Dropdown */}
      <Card>
        <CardHeader>
          <CardTitle>Dropdown</CardTitle>
          <CardDescription>With more content</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-2">
          {/* Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuItem>Profile</DropdownMenuItem>
                <DropdownMenuItem>Billing</DropdownMenuItem>
                <DropdownMenuSeparator />
              </DropdownMenuGroup>
              <DropdownMenuGroup>
                <DropdownMenuItem>Team</DropdownMenuItem>
                <DropdownMenuItem>Subscription</DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
          {/* Sous-menu */}
          <DropdownMenu>
            <DropdownMenuTrigger>Sous-menu</DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuItem>Profile</DropdownMenuItem>
                <DropdownMenuItem>Billing</DropdownMenuItem>
                <DropdownMenuSeparator />
              </DropdownMenuGroup>
              <DropdownMenuGroup>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>More Options</DropdownMenuSubTrigger>
                  <DropdownMenuSubContent>
                    <DropdownMenuItem>Option 1</DropdownMenuItem>
                    <DropdownMenuItem>Option 2</DropdownMenuItem>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
                <DropdownMenuItem>Team</DropdownMenuItem>
                <DropdownMenuItem>Subscription</DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </CardContent>
        <CardFooter>
          <p>Footer Content</p>
        </CardFooter>
      </Card>
      {/* Popover */}
      <Card>
        <CardHeader>
          <CardTitle>Popover</CardTitle>
          <CardDescription>Card Description</CardDescription>
          <CardAction>Card Action</CardAction>
        </CardHeader>
        <CardContent>
          <Popover>
            <PopoverTrigger>Open Popover</PopoverTrigger>
            <PopoverContent>
              <p>This is the content of the popover.</p>
            </PopoverContent>
          </Popover>
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>

      {/* Dialog, drawer */}
      <Card>
        <CardHeader>
          <CardTitle>Dialog, Drawer</CardTitle>
          <CardDescription>Card Description</CardDescription>
          <CardAction>Card Action</CardAction>
        </CardHeader>
        <CardContent>

          <ResponsiveDialog>
            <ResponsiveDialogTrigger>
              Open Responsive Dialog
            </ResponsiveDialogTrigger>
            <ResponsiveDialogContent>
              <ResponsiveDialogHeader>
                <ResponsiveDialogTitle>
                  Responsive Dialog Title
                </ResponsiveDialogTitle>
                <ResponsiveDialogDescription>
                  This is a responsive dialog description.
                </ResponsiveDialogDescription>
              </ResponsiveDialogHeader>
              <ResponsiveDialogBody>test</ResponsiveDialogBody>
              <ResponsiveDialogFooter>Footer Content</ResponsiveDialogFooter>
            </ResponsiveDialogContent>
          </ResponsiveDialog>
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>
      
      {/* sonner */}
      <Card>
        <CardHeader>
          <CardTitle>Sonner</CardTitle>
          <CardDescription>Card Description</CardDescription>
          <CardAction>Card Action</CardAction>
        </CardHeader>
        <CardContent>
          <TestToast />
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>

      {/* dataTable */}
      <Card className="col-span-3">
        <CardHeader>
          <CardTitle>Data Table</CardTitle>
          <CardDescription>With more content</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Column 1</TableHead>
                <TableHead>Column 2</TableHead>
                <TableHead>Column 3</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>1</TableCell>
                <TableCell>2</TableCell>
                <TableCell>3</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
        <CardFooter>
          <p>Footer Content</p>
        </CardFooter>
      </Card>
      <div className="col-span-3 h-16">
        <ComponentExample />
        </div>
    </div>
    </AppLayout>
  );
}
